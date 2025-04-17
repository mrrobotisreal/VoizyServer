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

func GetPopularPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("id")
	limitStr := r.URL.Query().Get("limit")
	daysStr := r.URL.Query().Get("days")

	fetchPopularPostsResponse, err := fetchPopularPosts(limitStr, daysStr)
	if err != nil {
		log.Println("Failed to fetch popular posts due to the following error: ", err)
		http.Error(w, "Failed to fetch popular posts", http.StatusInternalServerError)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Println("Failed to convert userIdStr (string) to userID (int64): ", err)
		http.Error(w, "Failed to convert param 'id'.", http.StatusInternalServerError)
		return
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		log.Println("Failed to convert limitString (string) to limit (int64): ", err)
		http.Error(w, "Failed to convert param 'limit'.", http.StatusInternalServerError)
		return
	}

	days, err := strconv.ParseInt(daysStr, 10, 64)
	if err != nil {
		log.Println("Failed to convert daysStr (string) to days (in64): ", err)
		http.Error(w, "Failed to convert param 'days'.", http.StatusInternalServerError)
		return
	}

	response, err := getPopularPostsInfo(fetchPopularPostsResponse.PostIDs, userID, limit, days, 1)
	if err != nil {
		log.Println("Failed to get popular posts info due to the following error: ", err)
		http.Error(w, "Failed to get popular posts.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func fetchPopularPosts(limit, days string) (models.FetchPopularPostsResponse, error) {
	baseURL := fmt.Sprintf("http://%s:%s/api/popular", os.Getenv("RECOMMENDATIONS_SERVICE_HOST"), os.Getenv("RECOMMENDATIONS_SERVICE_PORT"))

	params := url.Values{}
	params.Add("limit", limit)
	params.Add("days", days)

	fullURL := baseURL + "?" + params.Encode()

	client := &http.Client{}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return models.FetchPopularPostsResponse{}, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return models.FetchPopularPostsResponse{}, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.FetchPopularPostsResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var fetchPopularPostsResponse models.FetchPopularPostsResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&fetchPopularPostsResponse); err != nil {
		return models.FetchPopularPostsResponse{}, fmt.Errorf("error decoding response: %v", err)
	}

	return fetchPopularPostsResponse, nil
}

func getPopularPostsInfo(popularPosts []int64, userID, limit, days, page int64) (models.GetPopularPostsResponse, error) {
	log.Println("popularPosts: ", popularPosts)
	log.Println("popularPosts length: ", len(popularPosts))

	if len(popularPosts) == 0 {
		return models.GetPopularPostsResponse{
			PopularPosts: []models.PopularPost{},
			Limit:        limit,
			Page:         page,
		}, nil
	}

	offset := (page - 1) * limit
	placeholders := make([]string, len(popularPosts))
	args := make([]interface{}, len(popularPosts)+1)
	args[0] = userID
	for i, id := range popularPosts {
		placeholders[i] = "?"
		args[i+1] = id
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
		LEFT JOIN post_reactions pr_user ON p.post_id = pr_user.post_id AND pr_user.user_id = ?
		WHERE p.post_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return models.GetPopularPostsResponse{}, fmt.Errorf("failed to execute query for post information: %v", err)
	}
	defer rows.Close()

	var popularPostsList []models.PopularPost
	for rows.Next() {
		var p models.PopularPost
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
		popularPostsList = append(popularPostsList, p)
	}

	if err := rows.Err(); err != nil {
		return models.GetPopularPostsResponse{}, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return models.GetPopularPostsResponse{
		PopularPosts: popularPostsList,
		Limit:        limit,
		Page:         page,
	}, err
}
