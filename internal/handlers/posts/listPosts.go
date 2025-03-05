package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"VoizyServer/internal/util"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
)

func ListPostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	userIDString := r.URL.Query().Get("id")
	if userIDString == "" {
		http.Error(w, "Missing required param 'id'.", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		log.Println("Failed to convert userIDString (string) to userID (int64): ", err)
		http.Error(w, "Failed to convert param 'id'.", http.StatusInternalServerError)
		return
	}

	limitString := r.URL.Query().Get("limit")
	if limitString == "" {
		http.Error(w, "Missing required param 'limit'.", http.StatusBadRequest)
		return
	}
	limit, err := strconv.ParseInt(limitString, 10, 64)
	if err != nil {
		log.Println("Failed to convert limitString (string) to limit (int64): ", err)
		http.Error(w, "Failed to convert param 'limit'.", http.StatusInternalServerError)
		return
	}

	pageString := r.URL.Query().Get("page")
	if pageString == "" {
		http.Error(w, "Missing required param 'page'.", http.StatusBadRequest)
		return
	}
	page, err := strconv.ParseInt(pageString, 10, 64)
	if err != nil {
		log.Println("Failed to convert pageString (string) to page (int64): ", err)
		http.Error(w, "Failed to convert param 'page'.", http.StatusInternalServerError)
		return
	}

	response, err := listPosts(userID, limit, page)
	if err != nil {
		log.Println("Failed to list posts due to the following error: ", err)
		http.Error(w, "Failed to list posts.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listPosts(userID, limit, page int64) (models.ListPostsResponse, error) {
	offset := (page - 1) * limit

	var totalPosts int64
	countQuery := `
		SELECT COUNT(*)
		FROM posts
		WHERE user_id = ?
	`
	err := database.DB.QueryRow(countQuery, userID).Scan(&totalPosts)
	if err != nil {
		return models.ListPostsResponse{}, fmt.Errorf("failed to get totalPosts: %w", err)
	}

	selectQuery := `
		SELECT
			post_id,
			user_id,
			IFNULL(content_text, ''),
			DATE_FORMAT(created_at, '%Y-%m-%d %T') AS created_at,
			DATE_FORMAT(updated_at, '%Y-%m-%d %T') AS updated_at,
			IFNULL(location_name, ''),
			IFNULL(location_lat, 0),
			IFNULL(location_lng, 0),
			is_poll,
			IFNULL(poll_question, ''),
			IFNULL(poll_duration_type, ''),
			IFNULL(poll_duration_length, 0)
		FROM posts
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := database.DB.Query(selectQuery, userID, limit, offset)
	if err != nil {
		return models.ListPostsResponse{}, fmt.Errorf("failed to select posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	var listPosts []models.ListPost
	for rows.Next() {
		var p models.Post
		err := rows.Scan(
			&p.PostID,
			&p.UserID,
			&p.ContentText,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.LocationName,
			&p.LocationLat,
			&p.LocationLong,
			&p.IsPoll,
			&p.PollQuestion,
			&p.PollDurationType,
			&p.PollDurationLength,
		)
		if err != nil {
			log.Println("Scan rows error: ", err)
			continue
		}
		posts = append(posts, p)
		listPosts = append(listPosts, models.ListPost{
			PostID:             p.PostID,
			UserID:             p.UserID,
			ContentText:        util.SqlNullStringToPtr(p.ContentText),
			CreatedAt:          util.SqlNullTimeToPtr(p.CreatedAt),
			UpdatedAt:          util.SqlNullTimeToPtr(p.UpdatedAt),
			LocationName:       util.SqlNullStringToPtr(p.LocationName),
			LocationLat:        util.SqlNullFloat64ToPtr(p.LocationLat),
			LocationLong:       util.SqlNullFloat64ToPtr(p.LocationLong),
			IsPoll:             util.SqlNullBoolToPtr(p.IsPoll),
			PollQuestion:       util.SqlNullStringToPtr(p.PollQuestion),
			PollDurationType:   util.SqlNullStringToPtr(p.PollDurationType),
			PollDurationLength: util.SqlNullInt64ToPtr(p.PollDurationLength),
		})
	}

	if err = rows.Err(); err != nil {
		return models.ListPostsResponse{}, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	totalPages := int64(math.Ceil(float64(totalPosts) / float64(limit)))
	return models.ListPostsResponse{
		Posts:      listPosts,
		Limit:      limit,
		Page:       page,
		TotalPosts: totalPosts,
		TotalPages: totalPages,
	}, nil
}
