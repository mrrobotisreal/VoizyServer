package models

import "time"

type GetUserResponse struct {
	UserID    int64     `json:"userID"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
