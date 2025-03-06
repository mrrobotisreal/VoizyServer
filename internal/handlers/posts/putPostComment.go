package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"VoizyServer/internal/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func PutCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.PutCommentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := putComment(req)
	if err != nil {
		log.Println("Failed to put comment on post due to the following error: ", err)
		http.Error(w, "Failed to put comment on post.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(req.UserID, "comment_on_post", "comment", &response.CommentID, map[string]interface{}{
		"postID": req.PostID,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func putComment(req models.PutCommentRequest) (models.PutCommentResponse, error) {
	query := `
		INSERT INTO comments
		(post_id, user_id, content_text)
		VALUES
		(?, ?, ?)
	`
	result, err := database.DB.Exec(query, req.PostID, req.UserID, req.ContentText)
	if err != nil {
		return models.PutCommentResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to put commont on post due to the following error: %v", err),
		}, err
	}
	commentID, _ := result.LastInsertId()

	return models.PutCommentResponse{
		Success:   true,
		Message:   "Successfully put comment on post.",
		CommentID: commentID,
	}, nil
}
