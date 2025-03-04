package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/analytics"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func BatchTrackEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var events []models.AnalyticsEvent
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := batchTrackEvents(events)
	if err != nil {
		http.Error(w, "Error tracking events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func batchTrackEvents(events []models.AnalyticsEvent) (models.BatchTrackEventsResponse, error) {
	for _, ev := range events {
		if ev.UserID == 0 || ev.EventType == "" {
			continue
		}

		eventTime := time.Now()
		if ev.EventTime != nil {
			eventTime = *ev.EventTime
		}

		metaBytes, _ := json.Marshal(ev.Metadata)

		if ev.ObjectID != nil {
			query := `
				INSERT INTO analytics_events (user_id, event_type, object_type, object_id, event_time, metadata)
				VALUES (?, ?, ?, ?, ?, ?)
			`
			if _, err := database.DB.Exec(query, ev.UserID, ev.EventType, ev.ObjectType, *ev.ObjectID, eventTime, metaBytes); err != nil {
				log.Println("Error tracking event! ", ev.UserID, ev.EventType, ev.ObjectType, *ev.ObjectID, eventTime, metaBytes)
				return models.BatchTrackEventsResponse{
					Success: false,
				}, fmt.Errorf("error tracking event")
			}
		} else {
			query := `
				INSERT INTO analytics_events (user_id, event_type, object_type, event_time, metadata)
				VALUES (?, ?, ?, ?, ?)
			`
			if _, err := database.DB.Exec(query, ev.UserID, ev.EventType, ev.ObjectType, eventTime, metaBytes); err != nil {
				log.Println("Error tracking event! ", ev.UserID, ev.EventType, ev.ObjectType, eventTime, metaBytes)
				return models.BatchTrackEventsResponse{
					Success: false,
				}, fmt.Errorf("error tracking event")
			}
		}
	}

	return models.BatchTrackEventsResponse{
		Success: true,
	}, nil
}
