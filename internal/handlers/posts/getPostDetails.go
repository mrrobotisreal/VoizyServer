package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func GetPostDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	postIDString := r.URL.Query().Get("id")
	if postIDString == "" {
		http.Error(w, "Missing required param 'id'.", http.StatusBadRequest)
		return
	}
	postID, err := strconv.ParseInt(postIDString, 10, 64)
	if err != nil {
		log.Println("Failed to parse postIDString (string) to postID (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'id'.", http.StatusInternalServerError)
		return
	}

	response, err := getPostDetails(postID)
	if err != nil {
		log.Println("Failed to get post details due to the following error: ", err)
		http.Error(w, "Failed to get post details.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getPostDetails(postID int64) (models.GetPostDetailsResponse, error) {
	var reactions []models.Reaction
	queryReactions := `
		SELECT post_reaction_id, post_id, user_id, reaction_type, reacted_at
		FROM post_reactions
		WHERE post_id = ?
	`
	rows, err := database.DB.Query(queryReactions, postID)
	if err != nil {
		return models.GetPostDetailsResponse{}, fmt.Errorf("failed to get post reactions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Reaction
		err := rows.Scan(
			&r.ReactionID,
			&r.PostID,
			&r.UserID,
			&r.ReactionType,
			&r.ReactedAt,
		)
		if err != nil {
			log.Println("Scan rows error: ", err)
			continue
		}
		reactions = append(reactions, r)
	}
	if err = rows.Err(); err != nil {
		return models.GetPostDetailsResponse{}, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	var hashtagIDs []int64
	queryPostHashtags := `
		SELECT hashtag_id
		FROM post_hashtags
		WHERE post_id = ?
	`
	rows, err = database.DB.Query(queryPostHashtags, postID)
	if err != nil {
		return models.GetPostDetailsResponse{}, fmt.Errorf("failed to get post hashtags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var h int64
		err := rows.Scan(&h)
		if err != nil {
			log.Println("Scan rows error: ", err)
			continue
		}
		hashtagIDs = append(hashtagIDs, h)
	}
	if err = rows.Err(); err != nil {
		return models.GetPostDetailsResponse{}, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	var hashtags []string
	queryHashtags := `
		SELECT tag
		FROM hashtags
		WHERE hashtag_id = ?
	`
	for _, id := range hashtagIDs {
		var t string
		row := database.DB.QueryRow(queryHashtags, id)
		err := row.Scan(&t)
		if err != nil {
			log.Println("failed to query row: ", err)
			continue
		}
		hashtags = append(hashtags, t)
	}

	return models.GetPostDetailsResponse{
		Reactions: reactions,
		Hashtags:  hashtags,
	}, nil
}
