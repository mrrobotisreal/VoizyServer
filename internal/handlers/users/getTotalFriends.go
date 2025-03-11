package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func GetTotalFriendsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userIDString := r.URL.Query().Get("id")
	if userIDString == "" {
		http.Error(w, "Missing required param 'userID'.", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		log.Println("Error converting userIDString (string) to userID (int64): ", err)
		http.Error(w, "Failed to parse param 'id'. It should be an int >= 1.", http.StatusInternalServerError)
		return
	}

	response, err := getTotalFriends(userID)
	if err != nil {
		log.Println("Failed to getTotalFriends with the following error: ", err)
		http.Error(w, "failed to get total friends.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getTotalFriends(userID int64) (models.GetTotalFriendsResponse, error) {
	var response models.GetTotalFriendsResponse

	query := `
		SELECT COUNT(*)
		FROM friendships
		WHERE user_id = ? OR friend_id = ?
	`
	row := database.DB.QueryRow(query, userID, userID)
	err := row.Scan(
		&response.TotalFriends,
	)
	if err != nil {
		return models.GetTotalFriendsResponse{}, err
	}

	return response, nil
}
