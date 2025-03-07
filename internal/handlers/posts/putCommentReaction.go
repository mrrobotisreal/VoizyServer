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

func PutCommentReactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.PutCommentReactionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := putCommentReaction(req)
	if err != nil {
		log.Println("Failed to put reaction to comment due to the following error: ", err)
		http.Error(w, "Failed to put reaction to comment.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(req.UserID, "react_to_comment", "comment_reaction", &response.CommentReactionID, map[string]interface{}{
		"reaction_type": req.ReactionType,
		"comment_id":    req.CommentID,
		"post_id":       req.PostID,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func putCommentReaction(req models.PutCommentReactionRequest) (models.PutCommentReactionResponse, error) {
	query := `
		INSERT INTO comment_reactions
		(comment_id, user_id, reaction_type)
		VALUES
		(?, ?, ?)
	`
	result, err := database.DB.Exec(query, req.CommentID, req.UserID, req.ReactionType)
	if err != nil {
		return models.PutCommentReactionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to execute query due to the following error: %v", err),
		}, err
	}
	reactionID, _ := result.LastInsertId()

	return models.PutCommentReactionResponse{
		Success:           true,
		Message:           "Successfully put reaction on post",
		CommentReactionID: reactionID,
	}, nil
}
