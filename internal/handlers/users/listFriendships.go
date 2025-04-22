package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"VoizyServer/internal/util"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
)

func ListFriendshipsHandler(w http.ResponseWriter, r *http.Request) {
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

	response, err := listFriendships(userID, limit, page)
	if err != nil {
		log.Println("Failed to list friendships due to the following error: ", err)
		http.Error(w, "Failed to list friendships.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listFriendships(userID, limit, page int64) (models.ListFriendshipsResponse, error) {
	offset := (page - 1) * limit

	var totalFriends int64
	countQuery := `
		SELECT COUNT(*)
		FROM friendships
		WHERE user_id = ? OR friend_id = ?
	`
	err := database.DB.QueryRow(countQuery, userID, userID).Scan(&totalFriends)
	if err != nil {
		return models.ListFriendshipsResponse{}, fmt.Errorf("failed to get totalFriends: %w", err)
	}

	//selectQuery := `
	//	SELECT
	//		f.friendship_id,
	//		f.user_id,
	//		f.friend_id,
	//		f.status,
	//		f.created_at,
	//
	//		u.username AS friend_username,
	//
	//		up.first_name,
	//		up.last_name,
	//		up.preferred_name,
	//
	//		ui.image_url AS profile_pic_url
	//	FROM friendships f
	//	LEFT JOIN users u
	//		ON u.user_id = f.friend_id
	//	LEFT JOIN user_profiles up
	//		ON up.user_id = f.friend_id
	//	LEFT JOIN user_images ui
	//		ON ui.user_id = f.friend_id
	//		AND ui.is_profile_pic = 1
	//	WHERE f.status = 'accepted'
	//		AND (f.user_id = ? OR f.friend_id = ?)
	//	ORDER BY f.created_at DESC
	//	LIMIT ? OFFSET ?
	//`
	//rows, err := database.DB.Query(selectQuery, userID, userID, limit, offset)
	selectQuery := `
	  SELECT
		fq.friendship_id,
		fq.user_id,
		fq.friend_id,
		fq.status,
		fq.created_at,
	
		u.username         AS friend_username,
		up.first_name,
		up.last_name,
		up.preferred_name,
		ui.image_url       AS profile_pic_url
	
	  FROM (
		SELECT
		  f.*,
		  CASE
			WHEN f.user_id   = ? THEN f.friend_id
			ELSE f.user_id
		  END                    AS friend_user_id
		FROM friendships f
		WHERE f.status = 'accepted'
		  AND (f.user_id   = ? 
			OR f.friend_id = ?)
	  ) AS fq
	
	  LEFT JOIN users u
		ON u.user_id = fq.friend_user_id
	
	  LEFT JOIN user_profiles up
		ON up.user_id = fq.friend_user_id
	
	  LEFT JOIN user_images ui
		ON ui.user_id      = fq.friend_user_id
	   AND ui.is_profile_pic = 1
	
	  ORDER BY fq.created_at DESC
	  LIMIT  ? OFFSET ?
	`
	rows, err := database.DB.Query(
		selectQuery,
		userID, userID, userID,
		limit, offset,
	)
	if err != nil {
		return models.ListFriendshipsResponse{}, err
	}
	defer rows.Close()

	var friends []models.ListFriendship
	for rows.Next() {
		var f models.ListFriendship
		var uid, fid int64
		var friendUsername, firstName, lastName, preferredName, profilePicURL, status sql.NullString
		var createdAt sql.NullTime
		err := rows.Scan(
			&f.FriendshipID,
			&uid,
			&fid,
			&status,
			&createdAt,
			&friendUsername,
			&firstName,
			&lastName,
			&preferredName,
			&profilePicURL,
		)
		if err != nil {
			log.Println("Scan rows error: ", err)
			continue
		}
		if uid == userID {
			f.FriendID = uid
			f.UserID = fid
		}
		if uid != userID {
			f.FriendID = fid
			f.UserID = uid
		}
		f.Status = util.SqlNullStringToPtr(status)
		f.CreatedAt = util.SqlNullTimeToPtr(createdAt)
		f.FriendUsername = util.SqlNullStringToPtr(friendUsername)
		f.FirstName = util.SqlNullStringToPtr(firstName)
		f.LastName = util.SqlNullStringToPtr(lastName)
		f.PreferredName = util.SqlNullStringToPtr(preferredName)
		f.ProfilePicURL = util.SqlNullStringToPtr(profilePicURL)
		friends = append(friends, f)
	}
	if err = rows.Err(); err != nil {
		return models.ListFriendshipsResponse{}, err
	}
	totalPages := int64(math.Ceil(float64(totalFriends) / float64(limit)))

	return models.ListFriendshipsResponse{
		Friends:      friends,
		Limit:        limit,
		Page:         page,
		TotalFriends: totalFriends,
		TotalPages:   totalPages,
	}, nil
}
