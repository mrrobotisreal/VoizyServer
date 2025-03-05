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

func PutPostReactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.PutReactionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := putPostReaction(req)
	if err != nil {
		log.Println("Failed to put reaction to post due to the following error: ", err)
		http.Error(w, "Failed to put reaction to post.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(req.UserID, "reaction", "post", &req.PostID, map[string]interface{}{
		"reaction_type": req.ReactionType,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func putPostReaction(req models.PutReactionRequest) (models.PutReactionResponse, error) {
	query := `
		INSERT INTO post_reactions
		(post_id, user_id, reaction_type)
		VALUES
		(?, ?, ?)
	`
	result, err := database.DB.Exec(query, req.PostID, req.UserID, req.ReactionType)
	if err != nil {
		return models.PutReactionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to execute query due to the following error: %v", err),
		}, err
	}
	reactionID, _ := result.LastInsertId()

	return models.PutReactionResponse{
		Success:    true,
		Message:    "Successfully put reaction on post",
		ReactionID: reactionID,
	}, nil
}
