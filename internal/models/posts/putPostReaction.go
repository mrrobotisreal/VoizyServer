package models

type PutReactionRequest struct {
	PostID       int64  `json:"postID"`
	UserID       int64  `json:"userID"`
	ReactionType string `json:"reactionType"`
}

type PutReactionResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
	ReactionID int64  `json:"reactionID,omitempty"`
}
