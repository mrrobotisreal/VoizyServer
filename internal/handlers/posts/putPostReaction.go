package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"VoizyServer/internal/util"
	"database/sql"
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

	go util.TrackEvent(req.UserID, "react_to_post", "post_reaction", &response.ReactionID, map[string]interface{}{
		"reaction_type": req.ReactionType,
		"post_id":       req.PostID,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func putPostReaction(req models.PutReactionRequest) (models.PutReactionResponse, error) {
	var (
		existingID   int64
		existingType string
	)

	err := database.DB.
		QueryRow(`SELECT reaction_id, reaction_type
                  FROM post_reactions
                  WHERE post_id = ? AND user_id = ?`,
			req.PostID, req.UserID).
		Scan(&existingID, &existingType)
	if err != nil && err != sql.ErrNoRows {
		return models.PutReactionResponse{
			Success: false,
			Message: fmt.Sprintf("error checking existing reaction: %v", err),
		}, err
	}

	if err == sql.ErrNoRows {
		res, err := database.DB.Exec(
			`INSERT INTO post_reactions (post_id, user_id, reaction_type)
             VALUES (?, ?, ?)`,
			req.PostID, req.UserID, req.ReactionType,
		)
		if err != nil {
			return models.PutReactionResponse{
				Success: false,
				Message: fmt.Sprintf("error inserting reaction: %v", err),
			}, err
		}
		newID, _ := res.LastInsertId()
		return models.PutReactionResponse{
			Success:    true,
			Message:    "Reaction added",
			ReactionID: newID,
		}, nil
	}

	if existingType == req.ReactionType {
		_, err := database.DB.Exec(
			`DELETE FROM post_reactions WHERE reaction_id = ?`,
			existingID,
		)
		if err != nil {
			return models.PutReactionResponse{
				Success: false,
				Message: fmt.Sprintf("error removing reaction: %v", err),
			}, err
		}
		return models.PutReactionResponse{
			Success:    true,
			Message:    "Reaction removed",
			ReactionID: existingID,
		}, nil
	}

	_, err = database.DB.Exec(
		`UPDATE post_reactions
         SET reaction_type = ?
         WHERE reaction_id = ?`,
		req.ReactionType, existingID,
	)
	if err != nil {
		return models.PutReactionResponse{
			Success: false,
			Message: fmt.Sprintf("error updating reaction: %v", err),
		}, err
	}
	return models.PutReactionResponse{
		Success:    true,
		Message:    "Reaction updated",
		ReactionID: existingID,
	}, nil

	//query := `
	//	INSERT INTO post_reactions
	//	(post_id, user_id, reaction_type)
	//	VALUES
	//	(?, ?, ?)
	//`
	//result, err := database.DB.Exec(query, req.PostID, req.UserID, req.ReactionType)
	//if err != nil {
	//	return models.PutReactionResponse{
	//		Success: false,
	//		Message: fmt.Sprintf("Failed to execute query due to the following error: %v", err),
	//	}, err
	//}
	//reactionID, _ := result.LastInsertId()
	//
	//return models.PutReactionResponse{
	//	Success:    true,
	//	Message:    "Successfully put reaction on post",
	//	ReactionID: reactionID,
	//}, nil
}
