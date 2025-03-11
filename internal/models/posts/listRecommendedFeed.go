package models

import "time"

type RecommendedPost struct {
	PostID        int64     `json:"postID"`
	UserID        int64     `json:"userID"`
	ContentText   string    `json:"contentText"`
	CreatedAt     time.Time `json:"createdAt"`
	ReactionCount int64     `json:"reactionCount"`
	CommentCount  int64     `json:"commentCount"`
	Views         int64     `json:"views"`
	Impressions   int64     `json:"impressions"`
	Score         float64   `json:"score"`
}

type ListRecommendedFeedResponse struct {
	Posts      []RecommendedPost `json:"posts"`
	Limit      int64             `json:"limit"`
	Page       int64             `json:"page"`
	TotalPosts int64             `json:"totalPosts"`
	TotalPages int64             `json:"totalPages"`
}
