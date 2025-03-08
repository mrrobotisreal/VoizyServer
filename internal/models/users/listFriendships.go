package models

import "time"

type ListFriendship struct {
	FriendshipID int64      `json:"friendshipID"`
	UserID       int64      `json:"userID"`
	FriendID     int64      `json:"friendID"`
	Status       *string    `json:"status"`
	CreatedAt    *time.Time `json:"createdAt"`
}

type ListFriendshipsResponse struct {
	Friends      []ListFriendship `json:"friends"`
	Limit        int64            `json:"limit"`
	Page         int64            `json:"page"`
	TotalFriends int64            `json:"totalFriends"`
	TotalPages   int64            `json:"totalPages"`
}
