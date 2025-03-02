package models

import (
	"database/sql"
	"time"
)

type UserProfile struct {
	UserID          int64
	ProfileID       int64
	FirstName       sql.NullString
	LastName        sql.NullString
	PreferredName   sql.NullString
	BirthDate       sql.NullTime
	CityOfResidence sql.NullString
	PlaceOfWork     sql.NullString
	DateJoined      sql.NullTime
}

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
