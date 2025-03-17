package models

type UpdateCoverPicRequest struct {
	UserID		int64  `json:"userID"`
	ImageID 	int64  `json:"imageID"`
}

type UpdateCoverPicResponse struct {
	Success 		bool   `json:"success"`
	Message 		string `json:"message,omitempty"`
}
