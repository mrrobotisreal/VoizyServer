package models

type GetProfilePicResponse struct {
	ProfilePicURL *string `json:"profilePicURL"`
	CoverPicURL 	*string `json:"coverPicURL"`
}
