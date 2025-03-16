package models

type PutUserImagesRequest struct {
	UserID 		int64    `json:"userID"`
	Images []string `json:"Images"`
}

type PutUserImagesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
