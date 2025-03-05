package models

import "time"

type Reaction struct {
	ReactionID   int64     `json:"reactionID"`
	PostID       int64     `json:"postID"`
	UserID       int64     `json:"userID"`
	ReactionType string    `json:"reactionType"`
	ReactedAt    time.Time `json:"reactedAt"`
}
