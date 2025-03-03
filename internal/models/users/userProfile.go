package models

import "database/sql"

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
