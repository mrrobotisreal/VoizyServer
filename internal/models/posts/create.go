package models

type CreatePostRequest struct {
	UserID             int64    `json:"userID"`
	ContentText        string   `json:"contentText"`
	LocationName       string   `json:"locationName"`
	LocationLat        float64  `json:"locationLat"`
	LocationLong       float64  `json:"locationLong"`
	Images             []string `json:"images"`
	Hashtags           []string `json:"hashtags"`
	IsPoll             bool     `json:"isPoll"`
	PollQuestion       string   `json:"pollQuestion"`
	PollDurationType   string   `json:"pollDurationType"`
	PollDurationLength int64    `json:"pollDurationLength"`
	PollOptions        []string `json:"pollOptions"`
}

type CreatePostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	PostID  int64  `json:"postID,omitempty"`
}
