package models

type PutPostMediaRequest struct {
	PostID int64  `json:"postID"`
	Images []string `json:"images"`
}

type PutPostMediaResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	PostID  int64  `json:"postID,omitempty"`
}