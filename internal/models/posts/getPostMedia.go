package models

type GetMediaResponse struct {
	Images []string `json:"images"`
	Videos []string `json:"videos"`
}
