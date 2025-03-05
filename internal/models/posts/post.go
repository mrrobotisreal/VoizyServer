package models

import (
	"database/sql"
)

type Post struct {
	PostID             int64           `json:"postID"`
	UserID             int64           `json:"userID"`
	ContentText        sql.NullString  `json:"contentText"`
	CreatedAt          sql.NullTime    `json:"createdAt"`
	UpdatedAt          sql.NullTime    `json:"updatedAt"`
	LocationName       sql.NullString  `json:"locationName"`
	LocationLat        sql.NullFloat64 `json:"locationLat"`
	LocationLong       sql.NullFloat64 `json:"locationLong"`
	IsPoll             sql.NullBool    `json:"isPoll"`
	PollQuestion       sql.NullString  `json:"pollQuestion"`
	PollDurationType   sql.NullString  `json:"pollDurationType"`
	PollDurationLength sql.NullInt64   `json:"pollDurationLength"`
}
