package models

type PutPostViewRequest struct {
	PostID int64 `json:"postID"`
	UserID int64 `json:"userID"`
	Views  int64 `json:"views"`
}

type PutPostViewResponse struct {
	Success    bool  `json:"success"`
	PostID     int64 `json:"postID"`
	TotalViews int64 `json:"totalViews,omitempty"`
}
