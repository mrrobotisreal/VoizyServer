package models

type PutPostImpressionRequest struct {
	PostID      int64 `json:"postID"`
	UserID      int64 `json:"userID"`
	Impressions int64 `json:"impressions"`
}

type PutPostImpressionResponse struct {
	Success          bool  `json:"success"`
	PostID           int64 `json:"postID"`
	TotalImpressions int64 `json:"totalImpressions,omitempty"`
}
