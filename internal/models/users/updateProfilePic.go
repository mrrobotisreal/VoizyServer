package models

type UpdateProfilePicRequest struct {
	UserID  int64 `json:"userID"`
	ImageID int64 `json:"imageID"`
}

type UpdateProfilePicResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
