package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func GetFriendStatus(w http.ResponseWriter, r *http.Request) {
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

	friendIDString := q.Get("friend")
	if friendIDString == "" {
		http.Error(w, "Missing required param 'friend'.", http.StatusBadRequest)
		return
	}
	friendID, err := strconv.ParseInt(friendIDString, 10, 64)
	if err != nil {
		log.Println("Failed to parse friendIDString (string) to friendID (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'friend'.", http.StatusInternalServerError)
		return
	}

	response, err := getStatus(userID, friendID)
	if err != nil {
		log.Println("Failed to get friendship status due to the following error: ", err)
		http.Error(w, "Failed to get friendship status.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getStatus(userID, friendID int64) (models.GetFriendStatusResponse, error) {
	var status string

	const query = `
		SELECT IFNULL((
			SELECT status
			FROM friendships
			WHERE (user_id = ? AND friend_id = ?)
				OR (user_id = ? AND friend_id = ?)
			LIMIT 1
		), 'idle') AS status
	`
	err := database.DB.QueryRow(query,
		userID, friendID,
		friendID, userID,
	).Scan(&status)
	if err != nil {
		return models.GetFriendStatusResponse{}, err
	}

	return models.GetFriendStatusResponse{
		Status: status,
	}, nil
}
