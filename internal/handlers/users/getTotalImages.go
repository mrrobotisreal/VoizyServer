package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func GetTotalImages(w http.ResponseWriter, r *http.Request) {
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

	response, err := getTotalImages(userID)
	if err != nil {
		log.Println("Failed to get total images due to the following error: ", err)
		http.Error(w, "Failed to get total images.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getTotalImages(userID int64) (models.GetTotalImagesResponse, error) {
	var totalImages int64
	countQuery := `
		SELECT COUNT(*)
		FROM user_images
		WHERE user_id = ?
	`
	err := database.DB.QueryRow(countQuery, userID).Scan(&totalImages)
	if err != nil {
		return models.GetTotalImagesResponse{}, err
	}

	return models.GetTotalImagesResponse{
		TotalImages: totalImages,
	}, nil
}
