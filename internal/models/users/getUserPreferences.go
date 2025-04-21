package models

type GetUserPreferencesResponse struct {
	UserID                 int64  `json:"userID"`
	PrimaryColor           string `json:"primaryColor"`
	PrimaryAccent          string `json:"primaryAccent"`
	SecondaryColor         string `json:"secondaryColor"`
	SecondaryAccent        string `json:"secondaryAccent"`
	SongAutoplay           bool   `json:"songAutoplay"`
	ProfilePrimaryColor    string `json:"profilePrimaryColor"`
	ProfilePrimaryAccent   string `json:"profilePrimaryAccent"`
	ProfileSecondaryColor  string `json:"profileSecondaryColor"`
	ProfileSecondaryAccent string `json:"profileSecondaryAccent"`
	ProfileSongAutoplay    bool   `json:"profileSongAutoplay"`
}
