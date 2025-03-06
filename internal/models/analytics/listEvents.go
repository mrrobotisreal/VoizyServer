package models

import "time"

type ListEvent struct {
	EventID    int64                   `json:"eventID"`
	UserID     int64                   `json:"userID"`
	EventType  string                  `json:"eventType"`
	ObjectType *string                 `json:"objectType"`
	ObjectID   *int64                  `json:"objectID"`
	EventTime  *time.Time              `json:"eventTime"`
	Metadata   *map[string]interface{} `json:"metadata"`
}

type ListEventsResponse struct {
	Events      []ListEvent `json:"events"`
	Limit       int64       `json:"limit"`
	Page        int64       `json:"page"`
	TotalEvents int64       `json:"totalEvents"`
	TotalPages  int64       `json:"totalPages"`
}
