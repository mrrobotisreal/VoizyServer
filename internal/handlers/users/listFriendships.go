package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"VoizyServer/internal/util"
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
		WHERE user_id = ?
	`
	err := database.DB.QueryRow(countQuery, userID).Scan(&totalFriends)
	if err != nil {
		return models.ListFriendshipsResponse{}, fmt.Errorf("failed to get totalFriends: %w", err)
	}

	selectQuery := `
		SELECT
			friendship_id,
			user_id,
			friend_id,
			status,
			created_at
		FROM friendships
		WHERE user_id = ? AND status = 'accepted'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := database.DB.Query(selectQuery, userID, limit, offset)
	if err != nil {
		return models.ListFriendshipsResponse{}, err
	}
	var friends []models.ListFriendship
	for rows.Next() {
		var f models.Friendship
		err := rows.Scan(
			&f.FriendshipID,
			&f.UserID,
			&f.FriendID,
			&f.Status,
			&f.CreatedAt,
		)
		if err != nil {
			log.Println("Scan rows error: ", err)
			continue
		}
		friends = append(friends, models.ListFriendship{
			FriendshipID: f.FriendshipID,
			UserID:       f.UserID,
			FriendID:     f.FriendID,
			Status:       util.SqlNullStringToPtr(f.Status),
			CreatedAt:    util.SqlNullTimeToPtr(f.CreatedAt),
		})
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
