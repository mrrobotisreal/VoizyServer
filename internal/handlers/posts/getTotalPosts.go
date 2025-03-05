package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func GetTotalPostsHandler(w http.ResponseWriter, r *http.Request) {
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

	response, err := getTotalPosts(userID)
	if err != nil {
		log.Println("Failed to getTotalPosts with the following error: ", err)
		http.Error(w, "failed to get total posts.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getTotalPosts(userID int64) (models.GetTotalPostsResponse, error) {
	var response models.GetTotalPostsResponse

	query := `
		SELECT COUNT(*)
		FROM posts
		WHERE user_id = ?
	`
	row := database.DB.QueryRow(query, userID)
	err := row.Scan(
		&response.TotalPosts,
	)
	if err != nil {
		return models.GetTotalPostsResponse{}, err
	}

	return response, nil
}
