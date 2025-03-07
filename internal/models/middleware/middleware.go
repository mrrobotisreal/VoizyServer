package models

import "time"

type contextKey string

const (
	UsernameContextKey contextKey = "username"
	APIKeyContextKey   contextKey = "apiKey"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type APIKey struct {
	Key       string    `bson:"key" json:"key"`
	Created   time.Time `bson:"created" json:"created"`
	LastUsed  time.Time `bson:"lastUsed" json:"lastUsed"`
	ExpiresAt time.Time `bson:"expiresAt" json:"expiresAt"`
}
