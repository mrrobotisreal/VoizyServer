package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func PutUserImagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.PutUserImagesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := putUserImages(req)
	if err != nil {
		log.Println("Failed to put user images due to the following error: ", err)
		http.Error(w, "Failed to put user images.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func putUserImages(req models.PutUserImagesRequest) (models.PutUserImagesResponse, error) {
	query := `
		INSERT INTO user_images
		(user_id, image_url)
		VALUES
		(?, ?)
	`

	for _, image := range req.Images {
		if image == "" {
			continue
		}
		_, err := database.DB.Exec(query, req.UserID, image)
		if err != nil {
			return models.PutUserImagesResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to put user images due to the following error: %v", err),
			}, err
		}
	}
	return models.PutUserImagesResponse{
		Success:     true,
		Message:     "Successfully put user images.",
	}, nil
}
