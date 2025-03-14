package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func GetTotalCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	postIDString := r.URL.Query().Get("postID")
	if postIDString == "" {
		http.Error(w, "Missing required param 'postID'.", http.StatusBadRequest)
		return
	}
	postID, err := strconv.ParseInt(postIDString, 10, 64)
	if err != nil {
		log.Println("Failed to convert postIDString (string) to postID (int64): ", err)
		http.Error(w, "Failed to convert param 'postID'.", http.StatusInternalServerError)
		return
	}

	response, err := getTotalComments(postID)
	if err != nil {
		log.Println("Failed to get total comments due to the following error: ", err)
		http.Error(w, "Failed to get total comments.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getTotalComments(postID int64) (models.GetTotalCommentsResponse, error) {
	var response models.GetTotalCommentsResponse

	query := `
		SELECT COUNT(*)
		FROM comments
		WHERE post_id = ?
	`
	row := database.DB.QueryRow(query, postID)
	err := row.Scan(
		&response.TotalPosts,
	)
	if err != nil {
		return models.GetTotalCommentsResponse{}, err
	}

	return response, nil
}
