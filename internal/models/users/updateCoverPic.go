package models

type UpdateCoverPicRequest struct {
	UserID		int64  `json:"userID"`
	ImageID 	string `json:"imageID"`
}

type UpdateCoverPicResponse struct {
	Success 		bool   `json:"success"`
	Message 		string `json:"message,omitempty"`
}
