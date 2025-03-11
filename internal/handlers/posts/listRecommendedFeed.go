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
	"sort"
	"strconv"
	"strings"
	"time"
)

func ListRecommendedFeedHandler(w http.ResponseWriter, r *http.Request) {
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

	response, err := listRecommendedFeed(userID, limit, page)
	if err != nil {
		log.Println("Failed to list recommended posts due to the following error: ", err)
		http.Error(w, "Failed to list recommended posts.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(userID, "view_recommended_feed", "", nil, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listRecommendedFeed(userID, limit, page int64) (models.ListRecommendedFeedResponse, error) {
	friendIDs, err := fetchFriendsIDs(userID)
	if err != nil {
		return models.ListRecommendedFeedResponse{}, err
	}
	friendIDs = append(friendIDs, userID)

	posts, err := fetchCandidatePosts(friendIDs)
	if err != nil {
		return models.ListRecommendedFeedResponse{}, err
	}

	err = fillReactionAndCommentCounts(posts)
	if err != nil {
		return models.ListRecommendedFeedResponse{}, err
	}

	now := time.Now()
	for i, p := range posts {
		hoursOld := now.Sub(p.CreatedAt).Hours()
		timeDecay := 1.0 / (1.0 + hoursOld)

		popularity := math.Log(1.0 + float64(p.ReactionCount+p.CommentCount+p.Views+p.Impressions))

		friendFactor := 1.0
		if p.UserID == userID {
			friendFactor = 0.5
		}

		score := 2.0*timeDecay + 1.5*popularity + friendFactor

		posts[i].Score = score
	}

	sortPostsByScoreDesc(posts)

	totalPosts := len(posts)
	totalPages := int64(math.Ceil(float64(totalPosts) / float64(limit)))
	if len(posts) > int(limit) {
		posts = posts[:limit]
	}

	return models.ListRecommendedFeedResponse{
		Posts:      posts,
		Limit:      limit,
		Page:       page,
		TotalPosts: int64(totalPosts),
		TotalPages: totalPages,
	}, nil
}

func fetchFriendsIDs(userID int64) ([]int64, error) {
	query := `
		SELECT friend_id
		FROM friendships
		WHERE user_id = ?
			AND status = 'accepted'
	`
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friendIDs []int64
	for rows.Next() {
		var fID int64
		if err := rows.Scan(&fID); err != nil {
			return nil, err
		}
		friendIDs = append(friendIDs, fID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return friendIDs, nil
}

func fetchCandidatePosts(friendsIDs []int64) ([]models.RecommendedPost, error) {
	if len(friendsIDs) == 0 {
		return []models.RecommendedPost{}, nil
	}

	inClause := buildInClause(len(friendsIDs))
	query := fmt.Sprintf(`
		SELECT p.post_id, p.user_id, p.content_text, p.created_at, p.views, p.impressions
		FROM posts p
		WHERE p.created_at >= (NOW() - INTERVAL 7 DAY)
			AND p.user_id IN (%s)
	`, inClause)

	args := make([]interface{}, len(friendsIDs))
	for i, fID := range friendsIDs {
		args[i] = fID
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.RecommendedPost
	for rows.Next() {
		var p models.RecommendedPost
		if err := rows.Scan(&p.PostID, &p.UserID, &p.ContentText, &p.CreatedAt, &p.Views, &p.Impressions); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func fillReactionAndCommentCounts(posts []models.RecommendedPost) error {
	if len(posts) == 0 {
		return nil
	}

	postMap := make(map[int64]*models.RecommendedPost)
	var postIDs []int64
	for i := range posts {
		postMap[posts[i].PostID] = &posts[i]
		postIDs = append(postIDs, posts[i].PostID)
	}

	inClause := buildInClause(len(postIDs))
	queryReactions := fmt.Sprintf(`
		SELECT post_id, COUNT(*) as cnt
		FROM post_reactions
		WHERE post_id IN (%s)
		GROUP BY post_id
	`, inClause)
	argsR := make([]interface{}, len(postIDs))
	for i, id := range postIDs {
		argsR[i] = id
	}
	rRows, err := database.DB.Query(queryReactions, argsR...)
	if err != nil {
		return err
	}
	defer rRows.Close()

	for rRows.Next() {
		var pID int64
		var cnt int64
		if err := rRows.Scan(&pID, &cnt); err != nil {
			return err
		}
		if pf, ok := postMap[pID]; ok {
			pf.ReactionCount = cnt
		}
	}
	if err := rRows.Err(); err != nil {
		return err
	}

	queryComments := fmt.Sprintf(`
		SELECT post_id, COUNT(*) as cnt
		FROM comments
		WHERE post_id IN (%s)
		GROUP BY post_id
	`, inClause)
	argsC := make([]interface{}, len(postIDs))
	for i, id := range postIDs {
		argsC[i] = id
	}
	cRows, err := database.DB.Query(queryComments, argsC...)
	if err != nil {
		return err
	}
	defer cRows.Close()

	for cRows.Next() {
		var pID int64
		var cnt int64
		if err := cRows.Scan(&pID, &cnt); err != nil {
			return err
		}
		if pf, ok := postMap[pID]; ok {
			pf.CommentCount = cnt
		}
	}
	if err := cRows.Err(); err != nil {
		return err
	}

	return nil
}

func buildInClause(n int) string {
	placeholders := make([]string, n)
	for i := 0; i < n; i++ {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ",")
}

func sortPostsByScoreDesc(posts []models.RecommendedPost) {
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score > posts[j].Score
	})
}
