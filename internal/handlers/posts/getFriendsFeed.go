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
	"strconv"
)

func GetFriendFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("id")
	limitStr := r.URL.Query().Get("limit")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Println("Failed to convert userIDString (string) to userID (in64): ", err)
		http.Error(w, "Failed to convert param 'id'.", http.StatusInternalServerError)
		return
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		log.Println("Failed to convert limitString (string) to limit (int64): ", err)
		http.Error(w, "Failed to convert param 'limit'.", http.StatusInternalServerError)
		return
	}

	response, err := getFriendFeed(userID, limit, 1)
	if err != nil {
		log.Println("Failed to get friend posts due to the following error: ", err)
		http.Error(w, "Failed to get friend feed posts.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getFriendFeed(userID, limit, page int64) (models.GetFriendFeedResponse, error) {
	offset := (page - 1) * limit

	query := `
		SELECT
			p.post_id,
			p.user_id,
			p.to_user_id,
			p.original_post_id,
			p.impressions,
			(SELECT COUNT(*) FROM post_views pv WHERE pv.post_id = p.post_id) AS views,
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
			ui.image_url AS profile_pic_url,
			pr_user.reaction_type AS user_reaction,
			(SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.post_id) AS total_reactions,
			(SELECT COUNT(*) FROM comments c WHERE c.post_id = p.post_id) AS total_comments,
			(SELECT COUNT(*) FROM post_shares ps WHERE ps.post_id = p.post_id) AS total_post_shares
		FROM posts p
		JOIN (
			SELECT friend_id 
			FROM friendships 
			WHERE user_id = ? AND status = 'accepted'
			UNION
			SELECT user_id 
			FROM friendships 
			WHERE friend_id = ? AND status = 'accepted'
		) f ON p.user_id = f.friend_id
		LEFT JOIN users u ON p.user_id = u.user_id
		LEFT JOIN user_profiles up ON p.user_id = up.user_id
		LEFT JOIN user_images ui ON p.user_id = ui.user_id
		LEFT JOIN post_reactions pr_user ON p.post_id = pr_user.post_id AND pr_user.user_id = ?
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := database.DB.Query(query, userID, userID, userID, limit, offset)
	if err != nil {
		return models.GetFriendFeedResponse{}, fmt.Errorf("failed to execute query for friend posts: %v", err)
	}
	defer rows.Close()

	var friendPosts []models.FriendPost
	for rows.Next() {
		var p models.FriendPost
		var (
			originalPostID     sql.NullInt64
			contentText        sql.NullString
			createdAt          sql.NullTime
			updatedAt          sql.NullTime
			locationName       sql.NullString
			locationLat        sql.NullFloat64
			locationLong       sql.NullFloat64
			isPoll             sql.NullBool
			pollQuestion       sql.NullString
			pollDurationType   sql.NullString
			pollDurationLength sql.NullInt64
			username           sql.NullString
			firstName          sql.NullString
			lastName           sql.NullString
			preferredName      sql.NullString
			userReaction       sql.NullString
			profilePicURL      sql.NullString
			totalReactions     int64
			totalComments      int64
			totalPostShares    int64
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
			&profilePicURL,
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
		p.ProfilePicURL = util.SqlNullStringToPtr(profilePicURL)
		p.UserReaction = util.SqlNullStringToPtr(userReaction)
		p.TotalReactions = totalReactions
		p.TotalComments = totalComments
		p.TotalPostShares = totalPostShares
		friendPosts = append(friendPosts, p)
	}
	if err := rows.Err(); err != nil {
		return models.GetFriendFeedResponse{}, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return models.GetFriendFeedResponse{
		FriendPosts: friendPosts,
		Limit:       limit,
		Page:        page,
	}, nil
}
