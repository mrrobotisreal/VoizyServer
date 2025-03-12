package models

import "time"

type ListPost struct {
	PostID             int64      `json:"postID"`
	UserID             int64      `json:"userID"`
	ToUserID           int64      `json:"toUserID"`
	OriginalPostID     *int64     `json:"originalPostID"`
	FirstName					 *string    `json:"firstName"`
	LastName					 *string    `json:"lastName"`
	PreferredName			 *string    `json:"preferredName"`
	Username 				   *string    `json:"username"`
	Impressions        int64      `json:"impressions"`
	Views              int64      `json:"views"`
	ContentText        *string    `json:"contentText"`
	CreatedAt          *time.Time `json:"createdAt"`
	UpdatedAt          *time.Time `json:"updatedAt"`
	LocationName       *string    `json:"locationName"`
	LocationLat        *float64   `json:"locationLat"`
	LocationLong       *float64   `json:"locationLong"`
	IsPoll             *bool      `json:"isPoll"`
	PollQuestion       *string    `json:"pollQuestion"`
	PollDurationType   *string    `json:"pollDurationType"`
	PollDurationLength *int64     `json:"pollDurationLength"`
}

type ListPostsResponse struct {
	Posts      []ListPost `json:"posts"`
	Limit      int64      `json:"limit"`
	Page       int64      `json:"page"`
	TotalPosts int64      `json:"totalPosts"`
	TotalPages int64      `json:"totalPages"`
}
