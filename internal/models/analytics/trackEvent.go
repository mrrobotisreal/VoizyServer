package models

import "time"

type AnalyticsEvent struct {
	UserID     int64                  `json:"userID"`
	EventType  string                 `json:"eventType"`
	ObjectType string                 `json:"objectType,omitempty"`
	ObjectID   *int64                 `json:"objectID,omitempty"`
	EventTime  *time.Time             `json:"eventTime,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type BatchTrackEventsResponse struct {
	Success bool `json:"success"`
}
