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
	FBUID           string    `json:"FBUID"`
	ProfileID       int64     `json:"profileID"`
	APIKey          string    `json:"apiKey"`
	Token           string    `json:"token"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
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
