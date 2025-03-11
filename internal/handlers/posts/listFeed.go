package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"VoizyServer/internal/util"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
)

func ListFeedHandler(w http.ResponseWriter, r *http.Request) {
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
		log.Println("Failed to parse userIDString (string) to userID (int64) due to the following error: ", err)
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

	response, err := listFeed(limit, page)
	if err != nil {
		log.Println("Failed to list feed due to the following error: ", err)
		http.Error(w, "Failed to list feed.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(userID, "view_main_feed", "", nil, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listFeed(limit, page int64) (models.ListFeedResponse, error) {
	offset := (page - 1) * limit

	var totalPosts int64
	countQuery := `
		SELECT COUNT(*)
		FROM posts
		WHERE to_user_id = -1
	`
	err := database.DB.QueryRow(countQuery).Scan(&totalPosts)
	if err != nil {
		return models.ListFeedResponse{}, err
	}

	query := `
		SELECT
			post_id,
			user_id,
			to_user_id,
			original_post_id,
			impressions,
			views,
			content_text,
			created_at,
			updated_at,
			location_name,
			location_lat,
			location_lng,
			is_poll,
			poll_question,
			poll_duration_type,
			poll_duration_length
		FROM posts
		WHERE to_user_id = -1
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := database.DB.Query(query, limit, offset)
	if err != nil {
		return models.ListFeedResponse{}, err
	}
	defer rows.Close()

	var posts []models.ListPost
	for rows.Next() {
		var p models.Post
		err := rows.Scan(
			&p.PostID,
			&p.UserID,
			&p.ToUserID,
			&p.OriginalPostID,
			&p.Impressions,
			&p.Views,
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
		posts = append(posts, models.ListPost{
			PostID:             p.PostID,
			UserID:             p.UserID,
			OriginalPostID:     util.SqlNullInt64ToPtr(p.OriginalPostID),
			Impressions:        p.Impressions,
			Views:              p.Views,
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
	if err := rows.Err(); err != nil {
		return models.ListFeedResponse{}, err
	}

	totalPages := int64(math.Ceil(float64(totalPosts) / float64(limit)))

	return models.ListFeedResponse{
		Posts:      posts,
		Limit:      limit,
		Page:       page,
		TotalPosts: totalPosts,
		TotalPages: totalPages,
	}, nil
}
