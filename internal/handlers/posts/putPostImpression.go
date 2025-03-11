package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"VoizyServer/internal/util"
	"encoding/json"
	"log"
	"net/http"
)

func PutPostImpressionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.PutPostImpressionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := putPostImpression(req)
	if err != nil {
		log.Println("Failed to put post impressions due to the following error: ", err)
		http.Error(w, "Failed to put post impressions.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(req.UserID, "update_post_impressions", "post", &req.PostID, map[string]interface{}{
		"newImpressions":   req.Impressions,
		"totalImpressions": response.TotalImpressions,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func putPostImpression(req models.PutPostImpressionRequest) (models.PutPostImpressionResponse, error) {
	var currentImpressions int64
	selectQuery := `
		SELECT impressions
		FROM posts
		WHERE post_id = ?
	`
	err := database.DB.QueryRow(selectQuery, req.PostID).Scan(&currentImpressions)
	if err != nil {
		return models.PutPostImpressionResponse{
			Success: false,
			PostID:  req.PostID,
		}, err
	}

	totalImpressions := currentImpressions + req.Impressions
	updateQuery := `
		UPDATE posts
		SET impressions = ?
		WHERE post_id = ?
	`
	_, err = database.DB.Exec(updateQuery, totalImpressions, req.PostID)
	if err != nil {
		return models.PutPostImpressionResponse{
			Success: false,
			PostID:  req.PostID,
		}, err
	}

	return models.PutPostImpressionResponse{
		Success:          true,
		PostID:           req.PostID,
		TotalImpressions: totalImpressions,
	}, nil
}
