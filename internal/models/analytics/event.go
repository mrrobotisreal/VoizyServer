package models

import (
	"database/sql"
)

type Event struct {
	EventID    int64                            `json:"eventID"`
	UserID     int64                            `json:"userID"`
	EventType  string                           `json:"eventType"`
	ObjectType sql.NullString                   `json:"objectType"`
	ObjectID   sql.NullInt64                    `json:"objectID"`
	EventTime  sql.NullTime                     `json:"eventTime"`
	Metadata   sql.Null[map[string]interface{}] `json:"metadata"`
}
