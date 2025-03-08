package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"VoizyServer/internal/util"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
)

func ListFriendsInCommonHandler(w http.ResponseWriter, r *http.Request) {
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

	friendIDString := q.Get("friend_id")
	friendID, err := strconv.ParseInt(friendIDString, 10, 64)
	if err != nil {
		log.Println("Failed to parse friendIDString (string) to friendID (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'friend_id'.", http.StatusInternalServerError)
		return
	}

	limitString := q.Get("limit")
	limit, err := strconv.ParseInt(limitString, 10, 64)
	if err != nil {
		log.Println("Failed to parse limitString (string) to limit (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'limit'.", http.StatusInternalServerError)
		return
	}

	pageString := q.Get("page")
	page, err := strconv.ParseInt(pageString, 10, 64)
	if err != nil {
		log.Println("Failed to parse pageString (string) to page (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'page'.", http.StatusInternalServerError)
		return
	}

	response, err := listFriendsInCommon(userID, friendID, limit, page)
	if err != nil {
		log.Println("Failed to list friends in common due to the following error: ", err)
		http.Error(w, "Failed to list friends in common.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(userID, "view_common_friends", "", nil, map[string]interface{}{
		"friendID":             friendID,
		"totalFriendsInCommon": response.TotalFriendsInCommon,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listFriendsInCommon(userID, friendID, limit, page int64) (models.ListFriendsInCommonResponse, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM friendships f1
		JOIN friendships f2 ON f1.friend_id = f2.friend_id
		JOIN users u		ON u.user_id	= f1.friend_id
		WHERE f1.user_id = ?
		AND   f2.user_id = ?
		AND   f1.status  = 'accepted'
		AND   f2.status  = 'accepted'
	`
	var totalFriendsInCommon int64
	err := database.DB.QueryRow(countQuery, userID, friendID).Scan(&totalFriendsInCommon)
	if err != nil {
		return models.ListFriendsInCommonResponse{}, err
	}

	offset := (page - 1) * limit
	query := `
		SELECT u.user_id, u.username
		FROM friendships f1
		JOIN friendships f2 ON f1.friend_id = f2.friend_id
		JOIN users u		ON u.user_id	= f1.friend_id
		WHERE f1.user_id = ?
		AND   f2.user_id = ?
		AND   f1.status  = 'accepted'
		AND   f2.status  = 'accepted'
		ORDER BY u.user_id DESC
		LIMIT ? OFFSET ?
	`
	rows, err := database.DB.Query(query, userID, friendID, limit, offset)
	if err != nil {
		return models.ListFriendsInCommonResponse{}, err
	}
	defer rows.Close()

	var friendsInCommon []models.ListFriendInCommon
	for rows.Next() {
		var f models.ListFriendInCommon
		if err := rows.Scan(&f.UserID, &f.Username); err != nil {
			log.Println("Scan row error: ", err)
			continue
		}
		friendsInCommon = append(friendsInCommon, f)
	}
	if err := rows.Err(); err != nil {
		return models.ListFriendsInCommonResponse{}, err
	}
	totalPages := int64(math.Ceil(float64(totalFriendsInCommon) / float64(limit)))

	return models.ListFriendsInCommonResponse{
		FriendsInCommon:      friendsInCommon,
		Limit:                limit,
		Page:                 page,
		TotalFriendsInCommon: totalFriendsInCommon,
		TotalPages:           totalPages,
	}, nil
}
