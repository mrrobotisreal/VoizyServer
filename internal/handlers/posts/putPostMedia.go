package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func PutPostMediaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var request models.PutPostMediaRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := putPostMedia(request)
	if err != nil {
		log.Println("Failed to put post media due to the following error: ", err)
		http.Error(w, "Failed to put post media.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func putPostMedia(req models.PutPostMediaRequest) (models.PutPostMediaResponse, error) {
	query := `
		INSERT INTO post_media
		(post_id, media_url, media_type)
		VALUES
		(?, ?, ?)
	`

	for _, image := range req.Images {
		if image == "" {
			continue
		}
		_, err := database.DB.Exec(query, req.PostID, image, "image")
		if err != nil {
			return models.PutPostMediaResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to put post media due to the following error: %v", err),
			}, err
		}
	}

	return models.PutPostMediaResponse{
		Success:     true,
		Message:     "Successfully put post media.",
		PostID: req.PostID,
	}, nil
}
