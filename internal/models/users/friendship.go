package models

import "database/sql"

type Friendship struct {
	FriendshipID int64          `json:"friendshipID"`
	UserID       int64          `json:"userID"`
	FriendID     int64          `json:"friendID"`
	Status       sql.NullString `json:"status"`
	CreatedAt    sql.NullTime   `json:"createdAt"`
}
