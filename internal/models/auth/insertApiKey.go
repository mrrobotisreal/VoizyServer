package models

type InsertApiKeyRequest struct {
	UserID int64  `json:"userID"`
	APIKey string `json:"APIKey"`
}

type InsertApiKeyResponse struct {
	Success  bool  `json:"success"`
	APIKeyID int64 `json:"APIKeyID,omitempty"`
}
