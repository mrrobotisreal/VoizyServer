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
			p.post_id,
			p.user_id,
			p.to_user_id,
			p.original_post_id,
			p.impressions,
			p.views,
			p.content_text,
			p.created_at,
			p.updated_at,
			p.location_name,
			p.location_lat,
			p.location_lng,
			p.is_poll,
			p.poll_question,
			p.poll_duration_type,
			p.poll_duration_length,
			u.username,
			up.first_name,
			up.last_name,
			up.preferred_name,
			pr_user.reaction_type AS user_reaction,
			(SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.post_id) AS total_reactions,
			(SELECT COUNT(*) FROM comments c WHERE c.post_id = p.post_id) AS total_comments,
			(SELECT COUNT(*) FROM post_shares ps WHERE ps.post_id = p.post_id) AS total_post_shares
		FROM posts p
		LEFT JOIN users u
			ON u.user_id = p.user_id
		LEFT JOIN user_profiles up
			ON up.user_id = p.user_id
		LEFT JOIN post_reactions pr_user
			ON pr_user.post_id = p.post_id AND pr_user.user_id = ?
		WHERE (p.user_id = ? OR p.to_user_id = ?)
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := database.DB.Query(selectQuery, userID, userID, limit, offset)
	if err != nil {
		return models.ListPostsResponse{}, fmt.Errorf("failed to select posts: %w", err)
	}
	defer rows.Close()

	var listPosts []models.ListPost
	for rows.Next() {
		var p models.ListPost
		var (
			originalPostID		 sql.NullInt64
			contentText				 sql.NullString
			createdAt					 sql.NullTime
			updatedAt					 sql.NullTime
			locationName			 sql.NullString
			locationLat				 sql.NullFloat64
			locationLong			 sql.NullFloat64
			isPoll						 sql.NullBool
			pollQuestion			 sql.NullString
			pollDurationType	 sql.NullString
			pollDurationLength sql.NullInt64
			username 					 sql.NullString
			firstName					 sql.NullString
			lastName					 sql.NullString
			preferredName			 sql.NullString
			userReaction			 sql.NullString
			totalReactions		 int64
			totalComments			 int64
			totalPostShares		 int64
		)

		err := rows.Scan(
			&p.PostID,
			&p.UserID,
			&p.ToUserID,
			&originalPostID,
			&p.Impressions,
			&p.Views,
			&contentText,
			&createdAt,
			&updatedAt,
			&locationName,
			&locationLat,
			&locationLong,
			&isPoll,
			&pollQuestion,
			&pollDurationType,
			&pollDurationLength,
			&username,
			&firstName,
			&lastName,
			&preferredName,
			&userReaction,
			&totalReactions,
			&totalComments,
			&totalPostShares,
		)
		if err != nil {
			log.Println("Scan rows error: ", err)
			continue
		}
		p.OriginalPostID = util.SqlNullInt64ToPtr(originalPostID)
		p.ContentText = util.SqlNullStringToPtr(contentText)
		p.CreatedAt = util.SqlNullTimeToPtr(createdAt)
		p.UpdatedAt = util.SqlNullTimeToPtr(updatedAt)
		p.LocationName = util.SqlNullStringToPtr(locationName)
		p.LocationLat = util.SqlNullFloat64ToPtr(locationLat)
		p.LocationLong = util.SqlNullFloat64ToPtr(locationLong)
		p.IsPoll = util.SqlNullBoolToPtr(isPoll)
		p.PollQuestion = util.SqlNullStringToPtr(pollQuestion)
		p.PollDurationType = util.SqlNullStringToPtr(pollDurationType)
		p.PollDurationLength = util.SqlNullInt64ToPtr(pollDurationLength)
		p.Username = util.SqlNullStringToPtr(username)
		p.FirstName = util.SqlNullStringToPtr(firstName)
		p.LastName = util.SqlNullStringToPtr(lastName)
		p.PreferredName = util.SqlNullStringToPtr(preferredName)
		p.UserReaction = util.SqlNullStringToPtr(userReaction)
		p.TotalReactions = totalReactions
		p.TotalComments = totalComments
		p.TotalPostShares = totalPostShares
		listPosts = append(listPosts, p)
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
