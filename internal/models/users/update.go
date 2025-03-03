package models

type UpdateUserRequest struct {
	Username string `json:"username"`
}

type UpdateUserResponse struct {
	IsUpdateSuccessful bool `json:"isUpdateSuccessful"`
}
