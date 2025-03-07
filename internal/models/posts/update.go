package models

type UpdatePostRequest struct {
	PostID       int64    `json:"postID"`
	UserID       int64    `json:"userID"`
	ContentText  string   `json:"contentText"`
	LocationName string   `json:"locationName"`
	LocationLat  float64  `json:"locationLat"`
	LocationLong float64  `json:"locationLong"`
	Images       []string `json:"images"`
	Hashtags     []string `json:"hashtags"`
}

type UpdatePostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	PostID  int64  `json:"postID,omitempty"`
}
