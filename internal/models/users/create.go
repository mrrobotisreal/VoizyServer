package models

import "time"

type CreateUserRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	PreferredName string `json:"preferredName"`
	Username      string `json:"username"`
}

type CreateUserResponse struct {
	UserID          int64     `json:"userID"`
	ProfileID       int64     `json:"profileID"`
	APIKey          string    `json:"apiKey"`
	Email           string    `json:"email"`
	Username        string    `json:"username"`
	PreferredName   string    `json:"preferredName"`
	FirstName       string    `json:"firstName"`
	LastName        string    `json:"lastName"`
	BirthDate       time.Time `json:"birthDate"`
	CityOfResidence string    `json:"cityOfResidence"`
	PlaceOfWork     string    `json:"placeOfWork"`
	DateJoined      time.Time `json:"dateJoined"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
