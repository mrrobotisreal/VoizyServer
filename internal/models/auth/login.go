package models

import "time"

type User struct {
	UserID       int64     `json:"userID"`
	FBUID        *string   `json:"FBUID"`
	Email        string    `json:"email"`
	Phone        *string   `json:"phone"`
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
	FBUID             *string   `json:"FBUID"`
	Phone             *string   `json:"phone"`
	APIKey            string    `json:"apiKey"`
	Token             string    `json:"token"`
	Email             string    `json:"email"`
	Username          string    `json:"username"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
