package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"VoizyServer/internal/util"
	"encoding/json"
	"log"
	"net/http"
)

func PutPostViewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.PutPostViewRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := putPostView(req)
	if err != nil {
		log.Println("Failed to put post views due to the following error: ", err)
		http.Error(w, "Failed to put post views.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(req.UserID, "update_post_views", "post", &req.PostID, map[string]interface{}{
		"newViews":   req.Views,
		"totalViews": response.TotalViews,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func putPostView(req models.PutPostViewRequest) (models.PutPostViewResponse, error) {
	var currentViews int64
	selectQuery := `
		SELECT views
		FROM posts
		WHERE post_id = ?
	`
	err := database.DB.QueryRow(selectQuery, req.PostID).Scan(&currentViews)
	if err != nil {
		return models.PutPostViewResponse{
			Success: false,
			PostID:  req.PostID,
		}, err
	}

	totalViews := currentViews + req.Views
	updateQuery := `
		UPDATE posts
		SET views = ?
		WHERE post_id = ?
	`
	_, err = database.DB.Exec(updateQuery, totalViews, req.PostID)
	if err != nil {
		return models.PutPostViewResponse{
			Success: false,
			PostID:  req.PostID,
		}, err
	}

	return models.PutPostViewResponse{
		Success:    true,
		PostID:     req.PostID,
		TotalViews: totalViews,
	}, nil
}
