package models

type PutUserPreferencesRequest struct {
	UserID                 int64   `json:"userID"`
	PrimaryColor           *string `json:"primaryColor,omitempty"`
	PrimaryAccent          *string `json:"primaryAccent,omitempty"`
	SecondaryColor         *string `json:"secondaryColor,omitempty"`
	SecondaryAccent        *string `json:"secondaryAccent,omitempty"`
	SongAutoplay           *bool   `json:"songAutoplay,omitempty"`
	ProfilePrimaryColor    *string `json:"profilePrimaryColor,omitempty"`
	ProfilePrimaryAccent   *string `json:"profilePrimaryAccent,omitempty"`
	ProfileSecondaryColor  *string `json:"profileSecondaryColor,omitempty"`
	ProfileSecondaryAccent *string `json:"profileSecondaryAccent,omitempty"`
	ProfileSongAutoplay    *bool   `json:"profileSongAutoplay,omitempty"`
}

type PutUserPreferencesResponse struct {
	Success           bool   `json:"success"`
	Message           string `json:"message,omitempty"`
	UserPreferencesID int64  `json:"userPreferencesID,omitempty"`
}
