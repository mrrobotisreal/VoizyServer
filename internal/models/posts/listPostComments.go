package models

import "time"

type ListComment struct {
	CommentID   int64     `json:"commentID"`
	PostID      int64     `json:"postID"`
	UserID      int64     `json:"userID"`
	ContentText string    `json:"contentText"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ListCommentsResponse struct {
	Comments      []ListComment `json:"comments"`
	Limit         int64         `json:"limit"`
	Page          int64         `json:"page"`
	TotalComments int64         `json:"totalComments"`
	TotalPages    int64         `json:"totalPages"`
}
