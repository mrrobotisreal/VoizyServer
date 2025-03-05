package models

import "time"

type Post struct {
	PostID             int64     `json:"postID"`
	UserID             int64     `json:"userID"`
	ContentText        string    `json:"contentText"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	LocationName       *string   `json:"locationName"`
	LocationLat        *float64  `json:"locationLat"`
	LocationLong       *float64  `json:"locationLong"`
	IsPoll             bool      `json:"isPoll"`
	PollQuestion       *string   `json:"pollQuestion"`
	PollDurationType   *string   `json:"pollDurationType"`
	PollDurationLength *int64    `json:"pollDurationLength"`
}
