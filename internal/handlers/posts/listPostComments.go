package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
)

func ListPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
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

	limitString := r.URL.Query().Get("limit")
	if limitString == "" {
		http.Error(w, "Missing required param 'limit'.", http.StatusBadRequest)
		return
	}
	limit, err := strconv.ParseInt(limitString, 10, 64)
	if err != nil {
		log.Println("Failed to parse limitString (string) to limit (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'limit'.", http.StatusInternalServerError)
		return
	}

	pageString := r.URL.Query().Get("page")
	if pageString == "" {
		http.Error(w, "Missing required param 'page'.", http.StatusBadRequest)
		return
	}
	page, err := strconv.ParseInt(pageString, 10, 64)
	if err != nil {
		log.Println("Failed to parse pageString (string) to page (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'page'.", http.StatusInternalServerError)
		return
	}

	response, err := listPostComments(postID, limit, page)
	if err != nil {
		log.Println("Failed to list post comments due to the following error: ", err)
		http.Error(w, "Failed to list post comments.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listPostComments(postID, limit, page int64) (models.ListCommentsResponse, error) {
	offset := (page - 1) * limit

	var totalComments int64
	countQuery := `
		SELECT COUNT(*)
		FROM comments
		WHERE post_id = ?
	`
	err := database.DB.QueryRow(countQuery, postID).Scan(&totalComments)
	if err != nil {
		return models.ListCommentsResponse{}, fmt.Errorf("failed to get totalComments: %w", err)
	}

	selectQuery := `
		SELECT
			comment_id,
			post_id,
			user_id,
			content_text,
			created_at,
			updated_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := database.DB.Query(selectQuery, postID, limit, offset)
	if err != nil {
		return models.ListCommentsResponse{}, err
	}
	var comments []models.ListComment
	for rows.Next() {
		var c models.ListComment
		err := rows.Scan(
			&c.CommentID,
			&c.PostID,
			&c.UserID,
			&c.ContentText,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			log.Println("Scan rows error: ", err)
			continue
		}
		comments = append(comments, c)
	}
	if err = rows.Err(); err != nil {
		return models.ListCommentsResponse{}, err
	}
	totalPages := int64(math.Ceil(float64(totalComments) / float64(limit)))

	return models.ListCommentsResponse{
		Comments:      comments,
		Limit:         limit,
		Page:          page,
		TotalComments: totalComments,
		TotalPages:    totalPages,
	}, nil
}
