package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"VoizyServer/internal/util"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func ListPeopleYouMayKnow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	userIDString := q.Get("id")
	if userIDString == "" {
		http.Error(w, "Missing required param 'id'.", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		log.Println("Failed to parse userIDString (string) to userID (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'id'.", http.StatusInternalServerError)
		return
	}

	limitString := q.Get("limit")
	if limitString == "" {
		http.Error(w, "Missing required param 'limit'.", http.StatusBadRequest)
		return
	}
	limit, err := strconv.ParseInt(limitString, 10, 64)
	if err != nil {
		log.Println("Failed to parse limitString (string) to limit (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'limit'.", http.StatusInternalServerError)
		return
	}

	pageString := q.Get("page")
	if pageString == "" {
		http.Error(w, "Missing required param 'page'.", http.StatusBadRequest)
		return
	}
	page, err := strconv.ParseInt(pageString, 10, 64)
	if err != nil {
		log.Println("Failed to parse pageString (string) to page (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'page'.", http.StatusInternalServerError)
		return
	}

	response, err := listPeople(userID, limit, page)
	if err != nil {
		log.Println("Failed to list people you may know due to the following error: ", err)
		http.Error(w, "Failed to list people you may know.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listPeople(userID, limit, page int64) (models.ListPeopleYouMayKnowResponse, error) {
	offset := (page - 1) * limit

	const query = `
    WITH
    direct_friends AS (
      SELECT CASE WHEN user_id = ? THEN friend_id ELSE user_id END AS friend_id
      FROM friendships
      WHERE status = 'accepted'
        AND (user_id = ? OR friend_id = ?)
    ),
    friends_of_friends AS (
      SELECT
        CASE WHEN f.user_id = df.friend_id THEN f.friend_id ELSE f.user_id END AS fof_id,
        df.friend_id AS mutual_friend
      FROM friendships f
      JOIN direct_friends df
        ON (f.user_id = df.friend_id OR f.friend_id = df.friend_id)
      WHERE f.status = 'accepted'
    ),
    candidates AS (
      SELECT
        fof_id AS user_id,
        COUNT(DISTINCT mutual_friend) AS common_friend_count
      FROM friends_of_friends
      WHERE fof_id != ?                       -- never suggest yourself
        AND fof_id NOT IN (SELECT friend_id FROM direct_friends)
      GROUP BY fof_id
    ),
    interaction_scores AS (
      SELECT
        pr.user_id      AS user_id,
        COUNT(*)        AS reaction_score
      FROM post_reactions pr
      WHERE pr.post_id IN (
        SELECT post_id FROM posts WHERE user_id = ?
      )
        AND pr.user_id IN (SELECT user_id FROM candidates)
      GROUP BY pr.user_id
    ),
    total_scores AS (
      SELECT
        c.user_id,
        c.common_friend_count,
        COALESCE(i.reaction_score, 0) AS reaction_score,
        (c.common_friend_count + COALESCE(i.reaction_score,0)) AS score
      FROM candidates c
      LEFT JOIN interaction_scores i
        ON c.user_id = i.user_id
    ),
    user_city_parts AS (
      SELECT
        TRIM(LOWER(SUBSTRING_INDEX(city_of_residence, ',', 1)))  AS tok1,
        TRIM(LOWER(SUBSTRING_INDEX(city_of_residence, ',', -1))) AS tok2
      FROM user_profiles
      WHERE user_id = ?
    ),
    friend_rows AS (
      SELECT
        u.user_id,
        u.username,
        up.first_name,
        up.last_name,
        up.preferred_name,
        ui.image_url            AS profile_pic_url,
        up.city_of_residence,
        JSON_ARRAYAGG(fof.mutual_friend) AS friends_in_common,
        ts.score                AS score,
        1                        AS category
      FROM total_scores ts
      JOIN users u  ON u.user_id = ts.user_id
      LEFT JOIN user_profiles up
        ON up.user_id = u.user_id
      LEFT JOIN user_images ui
        ON ui.user_id = u.user_id AND ui.is_profile_pic = 1
      LEFT JOIN friends_of_friends fof
        ON fof.fof_id = u.user_id
      GROUP BY
        u.user_id, u.username,
        up.first_name, up.last_name, up.preferred_name,
        ui.image_url, up.city_of_residence
    ),
    city_rows AS (
      SELECT
        u.user_id,
        u.username,
        up2.first_name,
        up2.last_name,
        up2.preferred_name,
        ui2.image_url            AS profile_pic_url,
        up2.city_of_residence,
        JSON_ARRAY()             AS friends_in_common,
        1                         AS score,
        2                         AS category
      FROM users u
      JOIN user_profiles up2
        ON up2.user_id = u.user_id
      LEFT JOIN user_images ui2
        ON ui2.user_id = u.user_id AND ui2.is_profile_pic = 1
      JOIN user_city_parts ucp
      WHERE u.user_id != ?
        AND u.user_id NOT IN (SELECT friend_id FROM direct_friends)
        AND NOT EXISTS (
          SELECT 1 FROM total_scores ts WHERE ts.user_id = u.user_id
        )
        AND (
          LOWER(up2.city_of_residence) LIKE CONCAT('%', ucp.tok1, '%')
          OR LOWER(up2.city_of_residence) LIKE CONCAT('%', ucp.tok2, '%')
        )
    )
    SELECT
      user_id, username,
      first_name, last_name, preferred_name,
      profile_pic_url, city_of_residence,
      friends_in_common
    FROM (
      SELECT * FROM friend_rows
      UNION ALL
      SELECT * FROM city_rows
    ) AS combined
    ORDER BY category ASC, score DESC
    LIMIT ? OFFSET ?;`

	rows, err := database.DB.Query(query,
		// 1) direct_friends placeholders
		userID, userID, userID,
		// 2) exclude self in candidates
		userID,
		// 3) reaction filter
		userID,
		// 4) pull your city tokens
		userID,
		// 5) exclude self in city_rows
		userID,
		// 6) final pagination
		limit, offset,
	)
	if err != nil {
		return models.ListPeopleYouMayKnowResponse{}, err
	}
	defer rows.Close()

	//offset := (page - 1) * limit
	//
	//const query = `
	//WITH direct_friends AS (
	//  SELECT CASE WHEN user_id = ? THEN friend_id ELSE user_id END AS friend_id
	//  FROM friendships
	//  WHERE status = 'accepted' AND (user_id = ? OR friend_id = ?)
	//),
	//friends_of_friends AS (
	//  SELECT
	//    CASE WHEN f.user_id = df.friend_id THEN f.friend_id ELSE f.user_id END AS fof_id,
	//    df.friend_id AS mutual_friend
	//  FROM friendships f
	//  JOIN direct_friends df
	//    ON (f.user_id = df.friend_id OR f.friend_id = df.friend_id)
	//  WHERE f.status = 'accepted'
	//),
	//candidates AS (
	//  SELECT
	//    fof_id    AS user_id,
	//    COUNT(DISTINCT mutual_friend) AS common_friend_count
	//  FROM friends_of_friends
	//  WHERE fof_id != ?
	//    AND fof_id NOT IN (SELECT friend_id FROM direct_friends)
	//  GROUP BY fof_id
	//),
	//interaction_scores AS (
	//  SELECT
	//    pr.user_id     AS other_user,
	//    COUNT(*)       AS reaction_score
	//  FROM post_reactions pr
	//  WHERE pr.post_id IN (
	//    SELECT post_id FROM posts WHERE user_id = ?
	//  )
	//  AND pr.user_id IN (SELECT user_id FROM candidates)
	//  GROUP BY pr.user_id
	//),
	//total_scores AS (
	//  SELECT
	//    c.user_id,
	//    c.common_friend_count,
	//    COALESCE(i.reaction_score, 0) AS reaction_score,
	//    (c.common_friend_count + COALESCE(i.reaction_score, 0)) AS total_score
	//  FROM candidates c
	//  LEFT JOIN interaction_scores i
	//    ON c.user_id = i.other_user
	//)
	//SELECT
	//  u.user_id,
	//  u.username,
	//  up.first_name,
	//  up.last_name,
	//  up.preferred_name,
	//  ui.image_url        AS profile_pic_url,
	//  up.city_of_residence,
	//  JSON_ARRAYAGG(fof.mutual_friend) AS friends_in_common
	//FROM total_scores ts
	//JOIN users u  ON u.user_id = ts.user_id
	//LEFT
	//JOIN user_profiles up ON up.user_id = u.user_id
	//LEFT
	//JOIN user_images ui   ON ui.user_id = u.user_id AND ui.is_profile_pic = 1
	//LEFT
	//JOIN friends_of_friends fof ON fof.fof_id = u.user_id
	//GROUP BY
	//  u.user_id, u.username,
	//  up.first_name, up.last_name, up.preferred_name,
	//  ui.image_url, up.city_of_residence
	//ORDER BY ts.total_score DESC
	//LIMIT ? OFFSET ?;`
	//rows, err := database.DB.Query(query,
	//	// direct_friends x3
	//	userID, userID, userID,
	//	// exclude yourself
	//	userID,
	//	// posts→post_reactions filter
	//	userID,
	//	// paginate
	//	limit, offset,
	//)
	//if err != nil {
	//	return models.ListPeopleYouMayKnowResponse{}, err
	//}
	//defer rows.Close()

	var people []models.PersonYouMayKnow
	for rows.Next() {
		var (
			p                               models.PersonYouMayKnow
			uname, fn, ln, pn, picURL, city sql.NullString
			friendsJSON                     sql.NullString
		)
		if err := rows.Scan(
			&p.UserID,
			&uname,
			&fn, &ln, &pn,
			&picURL,
			&city,
			&friendsJSON,
		); err != nil {
			log.Println("scan error:", err)
			continue
		}

		p.Username = util.SqlNullStringToPtr(uname)
		p.FirstName = util.SqlNullStringToPtr(fn)
		p.LastName = util.SqlNullStringToPtr(ln)
		p.PreferredName = util.SqlNullStringToPtr(pn)
		p.ProfilePicURL = util.SqlNullStringToPtr(picURL)
		p.CityOfResidence = util.SqlNullStringToPtr(city)

		if friendsJSON.Valid {
			var fids []int64
			if err := json.Unmarshal([]byte(friendsJSON.String), &fids); err != nil {
				log.Println("unmarshal friends_in_common:", err)
			} else {
				p.FriendsInCommon = fids
			}
		}
		people = append(people, p)
	}
	if err := rows.Err(); err != nil {
		return models.ListPeopleYouMayKnowResponse{}, err
	}

	return models.ListPeopleYouMayKnowResponse{
		PeopleYouMayKnow: people,
		Limit:            limit,
		Page:             page,
	}, nil
}
