package models

import "time"

type UserImage struct {
	UserID           int64     `json:"userID"`
	ImageID 				 int64     `json:"imageID"`
	ImageURL         string    `json:"imageURL"`
	IsProfilePicture bool      `json:"isProfilePicture"`
	UploadedAt       time.Time `json:"uploadedAt"`
}

type ListImagesResponse struct {
	Images      []UserImage `json:"images"`
	Limit       int64       `json:"limit"`
	Page        int64       `json:"page"`
	TotalImages int64       `json:"totalImages"`
	TotalPages  int64       `json:"totalPages"`
}
