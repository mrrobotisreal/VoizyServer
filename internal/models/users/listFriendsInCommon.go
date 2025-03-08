package models

type ListFriendInCommon struct {
	UserID   int64  `json:"userID"`
	Username string `json:"username"`
}

type ListFriendsInCommonResponse struct {
	FriendsInCommon      []ListFriendInCommon `json:"friendsInCommon"`
	Limit                int64                `json:"limit"`
	Page                 int64                `json:"page"`
	TotalFriendsInCommon int64                `json:"totalFriendsInCommon"`
	TotalPages           int64                `json:"totalPages"`
}
