package models

import "time"

type Comment struct {
	CommentID   int64     `json:"commentID"`
	PostID      int64     `json:"postID"`
	UserID      int64     `json:"userID"`
	ContentText string    `json:"contentText"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
