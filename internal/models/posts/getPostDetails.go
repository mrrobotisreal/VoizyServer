package models

type GetPostDetailsResponse struct {
	Reactions []Reaction `json:"reactions"`
	Hashtags  []string   `json:"hashtags"`
}
