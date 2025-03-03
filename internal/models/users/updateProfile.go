package models

import "time"

type UpdateUserProfileRequest struct {
	FirstName       string    `json:"firstName"`
	LastName        string    `json:"lastName"`
	PreferredName   string    `json:"preferredName"`
	BirthDate       time.Time `json:"birthDate"`
	CityOfResidence string    `json:"cityOfResidence"`
	PlaceOfWork     string    `json:"placeOfWork"`
	DateJoined      time.Time `json:"dateJoined"`
}

type UpdateUserProfileResponse struct {
	IsUpdateSuccessful bool `json:"isUpdateSuccessful"`
}
