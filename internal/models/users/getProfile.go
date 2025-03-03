package models

import (
	"time"
)

type GetUserProfileResponse struct {
	UserID          int64      `json:"userID"`
	ProfileID       int64      `json:"profileID"`
	FirstName       *string    `json:"firstName"`
	LastName        *string    `json:"lastName"`
	PreferredName   *string    `json:"preferredName"`
	BirthDate       *time.Time `json:"birthDate"`
	CityOfResidence *string    `json:"cityOfResidence"`
	PlaceOfWork     *string    `json:"placeOfWork"`
	DateJoined      *time.Time `json:"dateJoined"`
}
