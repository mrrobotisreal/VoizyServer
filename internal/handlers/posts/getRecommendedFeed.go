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
	"net/url"
	"os"
	"strconv"
	"strings"
)

func GetRecommendedFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("id")
	limitStr := r.URL.Query().Get("limit")
	excludeSeenStr := r.URL.Query().Get("excludeSeen")

	recommendedPostsResponse, err := fetchRecommendations(userIDStr, limitStr, excludeSeenStr)
	if err != nil {
		log.Println("Failed to fetch recommended posts due to the following error: ", err)
		http.Error(w, "Failed to fetch recommended posts", http.StatusInternalServerError)
		return
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		log.Println("Failed to convert limitString (string) to limit (int64): ", err)
		http.Error(w, "Failed to convert param 'limit'.", http.StatusInternalServerError)
		return
	}

	response, err := getPostInfo(recommendedPostsResponse.Recommendations, limit, 1)
	if err != nil {
		log.Println("Failed to get posts info due to the following error: ", err)
		http.Error(w, "Failed to get recommended feed posts.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func fetchRecommendations(userID, limit, excludeSeen string) (models.ScoredPostsResponse, error) {
	baseURL := fmt.Sprintf("http://%s:%s/api/recommendations", os.Getenv("RECOMMENDATIONS_SERVICE_HOST"), os.Getenv("RECOMMENDATIONS_SERVICE_PORT"))
	// baseURL := `http://192.168.4.74:5000/api/recommendations`

	params := url.Values{}
	params.Add("user_id", userID)
	params.Add("limit", limit)
	params.Add("exclude_seen", excludeSeen)

	fullURL := baseURL + "?" + params.Encode()

	client := &http.Client{}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return models.ScoredPostsResponse{}, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return models.ScoredPostsResponse{}, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.ScoredPostsResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var fetchRecommendationsResponse models.ScoredPostsResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&fetchRecommendationsResponse); err != nil {
		return models.ScoredPostsResponse{}, fmt.Errorf("error decoding response: %v", err)
	}

	return fetchRecommendationsResponse, nil
}

func getPostInfo(recommendedPosts []models.ScoredPost, limit, page int64) (models.GetRecommendedFeedResponse, error) {
	log.Println("recommendedPosts: ", recommendedPosts)
	log.Println("recommendedPosts length: ", len(recommendedPosts))

	if len(recommendedPosts) == 0 {
		return models.GetRecommendedFeedResponse{}, nil
	}

	offset := (page - 1) * limit
	placeholders := make([]string, len(recommendedPosts))
	args := make([]interface{}, len(recommendedPosts))
	for i, id := range recommendedPosts {
		placeholders[i] = "?"
		args[i] = id.PostID
	}
	args = append(args, limit, offset)

	query := `
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
			ui.image_url AS profile_pic_url,
			pr_user.reaction_type AS user_reaction,
			(SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.post_id) AS total_reactions,
			(SELECT COUNT(*) FROM comments c WHERE c.post_id = p.post_id) AS total_comments,
			(SELECT COUNT(*) FROM post_shares ps WHERE ps.post_id = p.post_id) AS total_post_shares
		FROM posts p
		LEFT JOIN users u ON p.user_id = u.user_id
		LEFT JOIN user_profiles up ON p.user_id = up.user_id
		LEFT JOIN user_images ui ON p.user_id = ui.user_id
		LEFT JOIN post_reactions pr_user ON p.post_id = pr_user.post_id
		WHERE p.post_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return models.GetRecommendedFeedResponse{}, fmt.Errorf("failed to execute query for post information: %v", err)
	}
	defer rows.Close()

	var recommendedFeedPostsList []models.RecommendedFeedPost
	for rows.Next() {
		var p models.RecommendedFeedPost
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
		recommendedFeedPostsList = append(recommendedFeedPostsList, p)
	}

	if err := rows.Err(); err != nil {
		return models.GetRecommendedFeedResponse{}, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return models.GetRecommendedFeedResponse{
		RecommendedFeedPosts: recommendedFeedPostsList,
		Limit:                limit,
		Page:                 page,
	}, err
}
