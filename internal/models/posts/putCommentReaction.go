package models

type PutCommentReactionRequest struct {
	CommentID    int64  `json:"commentID"`
	PostID       int64  `json:"postID"`
	UserID       int64  `json:"userID"`
	ReactionType string `json:"reactionType"`
}

type PutCommentReactionResponse struct {
	Success           bool   `json:"success"`
	Message           string `json:"message,omitempty"`
	CommentReactionID int64  `json:"commentReactionID,omitempty"`
}
