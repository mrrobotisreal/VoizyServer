package handlers

import (
	database "VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func GetProfilePicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	userIDString := r.URL.Query().Get("id")
	if userIDString == "" {
		http.Error(w, "Missing required param 'id'.", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		log.Println("Failed to convert userIDString (string) to userID (int64): ", err)
		http.Error(w, "Failed to convert param 'id'.", http.StatusInternalServerError)
		return
	}

	response, err := getProfilePic(userID)
	if err != nil {
		log.Println("Failed to get profile pic due to the following error: ", err)
		http.Error(w, "Failed to get profile pic.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getProfilePic(userID int64) (models.GetProfilePicResponse, error) {
	var response models.GetProfilePicResponse
	query := `SELECT image_url FROM user_images WHERE user_id = ? AND is_profile_picture = 1 LIMIT 1`
	err := database.DB.QueryRow(query, userID).Scan(&response.ProfilePicURL)
	if err != nil {
		return models.GetProfilePicResponse{}, err
	}

	return response, nil
}