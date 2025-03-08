package models

type CreateFriendRequestRequest struct {
	UserID   int64 `json:"userID"`
	FriendID int64 `json:"friendID"`
}

type CreateFriendRequestResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
	FriendshipID int64  `json:"friendshipID,omitempty"`
}
