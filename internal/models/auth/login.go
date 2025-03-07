package models

import "time"

type User struct {
	UserID       int64     `json:"userID"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	Salt         string    `json:"salt"`
	PasswordHash string    `json:"passwordHash"`
	APIKey       string    `json:"APIKey"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	IsPasswordCorrect bool      `json:"isPasswordCorrect"`
	UserID            int64     `json:"userID"`
	APIKey            string    `json:"apiKey"`
	Token             string    `json:"token"`
	Email             string    `json:"email"`
	Username          string    `json:"username"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
