package models

import "time"

type contextKey string

const (
	UserIDContextKey contextKey = "userID"
	APIKeyContextKey contextKey = "apiKey"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type APIKey struct {
	Key        string    `json:"key"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
