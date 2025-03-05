package models

type PutCommentRequest struct {
	PostID      int64  `json:"postID"`
	UserID      int64  `json:"userID"`
	ContentText string `json:"contentText"`
}

type PutCommentResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	CommentID int64  `json:"commentID,omitempty"`
}
