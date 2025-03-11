package models

type ListFeedResponse struct {
	Posts      []ListPost `json:"posts"`
	Limit      int64      `json:"limit"`
	Page       int64      `json:"page"`
	TotalPosts int64      `json:"totalPosts"`
	TotalPages int64      `json:"totalPages"`
}
