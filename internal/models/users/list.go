package models

type ListUsersResponse struct {
	Profiles []UserProfile `json:"profiles"`
}
