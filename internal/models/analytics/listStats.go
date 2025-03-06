package models

type StatSummary struct {
	GroupLabel string `json:"groupLabel"`
	GroupValue string `json:"groupValue"`
	Count      int64  `json:"count"`
}

type ListStatsResponse struct {
	Stats []StatSummary `json:"stats"`
}
