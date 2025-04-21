package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"VoizyServer/internal/util"
	sql2 "database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func SearchPeople(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	var req models.SearchPeopleRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := search(req.Query, userID, limit, page)
	if err != nil {
		log.Println("Failed to search people due to the following error: ", err)
		http.Error(w, "Failed to search people.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func search(searchQuery string, userID, limit, page int64) (models.SearchPeopleResponse, error) {
	offset := (page - 1) * limit

	const sql = `
    WITH
      direct_friends AS (
        SELECT CASE WHEN user_id = ? THEN friend_id ELSE user_id END AS friend_id
        FROM friendships
        WHERE status = 'accepted'
          AND (? IN (user_id, friend_id))
      ),
      friends_of_friends AS (
        SELECT
          CASE WHEN f.user_id = df.friend_id THEN f.friend_id ELSE f.user_id END AS fof_id,
          df.friend_id AS mutual_friend
        FROM friendships f
        JOIN direct_friends df
          ON f.status = 'accepted'
         AND (f.user_id = df.friend_id OR f.friend_id = df.friend_id)
      ),
      user_city_parts AS (
        SELECT
          TRIM(LOWER(SUBSTRING_INDEX(city_of_residence, ',', 1)))  AS tok1,
          TRIM(LOWER(SUBSTRING_INDEX(city_of_residence, ',', -1))) AS tok2
        FROM user_profiles
        WHERE user_id = ?
      ),
      matches AS (
        SELECT
          u.user_id,
          u.username,
          up.first_name,
          up.last_name,
          up.preferred_name,
          up.city_of_residence,
          ui.image_url AS profile_pic_url,
          CONCAT_WS(' ', COALESCE(up.first_name,''), COALESCE(up.last_name,''))     AS fn_ln,
          CONCAT_WS(' ', COALESCE(up.preferred_name,''), COALESCE(up.last_name,'')) AS pn_ln
        FROM users u
        LEFT JOIN user_profiles up ON up.user_id = u.user_id
        LEFT JOIN user_images   ui ON ui.user_id = u.user_id AND ui.is_profile_pic = 1
        WHERE u.user_id != ?
          AND (
            LOWER(u.username) LIKE CONCAT('%', LOWER(?), '%')
            OR LOWER(up.first_name)     LIKE CONCAT('%', LOWER(?), '%')
            OR LOWER(up.last_name)      LIKE CONCAT('%', LOWER(?), '%')
            OR LOWER(up.preferred_name) LIKE CONCAT('%', LOWER(?), '%')
            OR LOWER(CONCAT_WS(' ',
               COALESCE(up.first_name, ''),
               COALESCE(up.last_name, '')
           ))
           LIKE CONCAT('%', LOWER(?), '%')
           OR LOWER(CONCAT_WS(' ',
               COALESCE(up.preferred_name, ''),
               COALESCE(up.last_name, '')
           ))
           LIKE CONCAT('%', LOWER(?), '%')
          )
      )
    SELECT
      m.user_id,
      m.username,
      m.first_name,
      m.last_name,
      m.preferred_name,
      m.profile_pic_url,
      m.city_of_residence,
      COALESCE(
        (SELECT JSON_ARRAYAGG(fof.mutual_friend)
         FROM friends_of_friends fof
         WHERE fof.fof_id = m.user_id),
        JSON_ARRAY()
      ) AS friends_in_common,
      CASE
        WHEN m.user_id IN (SELECT friend_id FROM direct_friends) THEN 1
        WHEN m.user_id IN (SELECT fof_id    FROM friends_of_friends) THEN 2
        WHEN EXISTS (
          SELECT 1 FROM user_city_parts ucp
          WHERE LOWER(m.city_of_residence) LIKE CONCAT('%', ucp.tok1, '%')
             OR LOWER(m.city_of_residence) LIKE CONCAT('%', ucp.tok2, '%')
        ) THEN 3
        ELSE 4
      END AS category,
      LEAST(
        IFNULL(INSTR(LOWER(m.username), LOWER(?)), 999),
        IFNULL(INSTR(LOWER(m.first_name), LOWER(?)), 999),
        IFNULL(INSTR(LOWER(m.last_name), LOWER(?)), 999),
        IFNULL(INSTR(LOWER(m.preferred_name), LOWER(?)), 999),
        IFNULL(INSTR(LOWER(m.fn_ln), LOWER(?)), 999),
        IFNULL(INSTR(LOWER(m.pn_ln), LOWER(?)), 999)
      ) AS match_rank
    FROM matches m
    ORDER BY category ASC, match_rank ASC
    LIMIT ? OFFSET ?;`

	args := []interface{}{
		userID,
		userID,
		userID,
		userID,
		searchQuery, searchQuery,
		searchQuery, searchQuery,
		searchQuery, searchQuery,
		searchQuery, searchQuery,
		searchQuery, searchQuery,
		searchQuery, searchQuery,
		limit, offset,
	}

	rows, err := database.DB.Query(sql, args...)
	if err != nil {
		return models.SearchPeopleResponse{}, err
	}
	defer rows.Close()

	var results []models.SearchPerson
	for rows.Next() {
		var (
			p                                            models.SearchPerson
			uname, fn, ln, pn, picURL, city, friendsJSON sql2.NullString
		)

		if err := rows.Scan(
			&p.UserID,
			&uname,
			&fn, &ln, &pn,
			&picURL,
			&city,
			&friendsJSON,
			new(int), new(int),
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

		results = append(results, p)
	}
	if err := rows.Err(); err != nil {
		return models.SearchPeopleResponse{}, err
	}

	return models.SearchPeopleResponse{
		Results: results,
		Limit:   limit,
		Page:    page,
	}, nil
}
