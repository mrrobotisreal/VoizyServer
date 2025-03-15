package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"VoizyServer/internal/util"
	"database/sql"
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
			c.comment_id,
			c.post_id,
			c.user_id,
			c.content_text,
			c.created_at,
			c.updated_at,
			u.username,
			up.first_name,
			up.last_name,
			up.preferred_name,
			ui.image_url AS profile_picture,
			GROUP_CONCAT(DISTINCT cr.reaction_type ORDER BY cr.reacted_at SEPARATOR ', ') AS distinct_reactions,
			COUNT(cr.comment_reaction_id) AS reaction_count
		FROM comments c
		LEFT JOIN users u
			ON c.user_id = u.user_id
		LEFT JOIN user_profiles up
			ON u.user_id = up.user_id
		LEFT JOIN user_images ui
			ON u.user_id = ui.user_id
			AND ui.is_profile_pic = 1
		LEFT JOIN comment_reactions cr
			ON c.comment_id = cr.comment_id
		WHERE c.post_id = ?
		GROUP BY c.comment_id, c.post_id, c.user_id, c.content_text, c.created_at, c.updated_at,
						 u.username, up.first_name, up.last_name, up.preferred_name, ui.image_url
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := database.DB.Query(selectQuery, postID, limit, offset)
	if err != nil {
		return models.ListCommentsResponse{}, err
	}
	var comments []models.ListComment
	for rows.Next() {
		var c models.ListComment
		var username, firstName, lastName, preferredName, profilePicURL, reactions sql.NullString
		var reactionCount int64
		err := rows.Scan(
			&c.CommentID,
			&c.PostID,
			&c.UserID,
			&c.ContentText,
			&c.CreatedAt,
			&c.UpdatedAt,
			&username,
			&firstName,
			&lastName,
			&preferredName,
			&profilePicURL,
			&reactions,
			&reactionCount,
		)
		if err != nil {
			log.Println("Scan rows error: ", err)
			continue
		}
		c.username = util.SqlNullStringToPointer(username)
		c.firstName = util.SqlNullStringToPointer(firstName)
		c.lastName = util.SqlNullStringToPointer(lastName)
		c.preferredName = util.SqlNullStringToPointer(preferredName)
		c.ProfilePicURL = util.SqlNullStringToPointer(profilePicURL)
		if reactions.Valid && reactions.String != "" {
			reactionsSlice := util.SplitAndTrim(reactions.String, ",")
			c.Reactions = reactionsSlice
		} else {
			c.Reactions = []string{}
		}
		c.ReactionCount = reactionCount
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
