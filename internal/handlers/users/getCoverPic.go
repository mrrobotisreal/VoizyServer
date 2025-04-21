package handlers

import (
	database "VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func GetCoverPicHandler(w http.ResponseWriter, r *http.Request) {
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

	response, err := getCoverPic(userID)
	if err != nil {
		log.Println("Failed to get cover pic due to the following error: ", err)
		http.Error(w, "Failed to get cover pic.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getCoverPic(userID int64) (models.GetCoverPicResponse, error) {
	var response models.GetCoverPicResponse
	query := `SELECT image_url FROM user_images WHERE user_id = ? AND is_cover_pic = 1 LIMIT 1`
	err := database.DB.QueryRow(query, userID).Scan(&response.CoverPicURL)
	if err != nil {
		if err == sql.ErrNoRows {
			response.CoverPicURL = ""
			return response, nil
		}
		return models.GetCoverPicResponse{}, err
	}

	return response, nil
}
