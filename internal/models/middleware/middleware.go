package models

import "time"

type APIKey struct {
	Key       string    `bson:"key" json:"key"`
	Created   time.Time `bson:"created" json:"created"`
	LastUsed  time.Time `bson:"lastUsed" json:"lastUsed"`
	ExpiresAt time.Time `bson:"expiresAt" json:"expiresAt"`
}
