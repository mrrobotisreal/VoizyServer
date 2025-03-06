package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/analytics"
	"VoizyServer/internal/util"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ListEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	userIDString := q.Get("id")
	if userIDString == "" {
		http.Error(w, "Missing required param 'id'.", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		log.Println("Failed to parse userIDString (string) to userID (int64) due to the following reason: ", err)
		http.Error(w, "Failed to parse param 'id'.", http.StatusInternalServerError)
		return
	}

	eventType := q.Get("event_type")
	if eventType == "" {
		http.Error(w, "Missing required param 'eventType'.", http.StatusBadRequest)
		return
	}

	objectType := q.Get("object_type")

	var startTime time.Time
	startTimeString := q.Get("start_time")
	if startTimeString == "" {
		distantPast := time.Unix(-10000000000, 0)
		timeString := distantPast.Format(time.RFC3339)
		startTime, err = time.Parse(time.RFC3339, timeString)
	} else {
		startTime, err = time.Parse(time.RFC3339, startTimeString)
	}
	if err != nil {
		log.Println("Failed to parse startTimeString (string) to startTime (time.Time) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'startTime'.", http.StatusInternalServerError)
		return
	}

	var endTime time.Time
	endTimeString := q.Get("end_time")
	if endTimeString == "" {
		now := time.Now()
		timeString := now.Format(time.RFC3339)
		endTime, err = time.Parse(time.RFC3339, timeString)
	} else {
		endTime, err = time.Parse(time.RFC3339, endTimeString)
	}
	if err != nil {
		log.Println("Failed to parse endTimeString (string) to endTime (time.Time) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'endTime'.", http.StatusInternalServerError)
		return
	}

	limitString := q.Get("limit")
	if limitString == "" {
		http.Error(w, "Missing required param 'limit'.", http.StatusBadRequest)
		return
	}
	limit, err := strconv.ParseInt(limitString, 10, 64)
	if err != nil {
		log.Println("Failed to parse limitString (string) to limit (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'limit'.", http.StatusInternalServerError)
		return
	}

	pageString := q.Get("page")
	if pageString == "" {
		http.Error(w, "Missing required param 'page'.", http.StatusBadRequest)
		return
	}
	page, err := strconv.ParseInt(pageString, 10, 64)
	if err != nil {
		log.Println("Failed to parse pageString (string) to page (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'page'.", http.StatusInternalServerError)
		return
	}

	response, err := listEvents(userID, eventType, objectType, startTime, endTime, limit, page)
	if err != nil {
		log.Println("Failed to list events due to the following error: ", err)
		http.Error(w, "Failed to list events.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listEvents(userID int64, eventType, objectType string, startTime, endTime time.Time, limit, page int64) (models.ListEventsResponse, error) {
	whereClauses := []string{"user_id = ? AND event_type = ?"}
	args := []interface{}{userID, eventType}
	if objectType != "" {
		whereClauses = append(whereClauses, "object_type = ? AND event_time >= ? AND event_time <= ?")
		args = append(args, objectType, startTime, endTime)
	} else {
		whereClauses = append(whereClauses, "event_time >= ? AND event_time <= ?")
		args = append(args, startTime, endTime)
	}
	whereSQL := strings.Join(whereClauses, " AND ")

	var totalEvents int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM analytics_events WHERE %s`, whereSQL)
	if err := database.DB.QueryRow(countQuery, args...).Scan(&totalEvents); err != nil {
		return models.ListEventsResponse{}, err
	}

	selectQuery := fmt.Sprintf(`
		SELECT
			event_id,
			user_id,
			event_type,
			object_type,
			object_id,
			event_time,
			metadata
		FROM analytics_events
		WHERE %s
		ORDER BY event_time DESC
		LIMIT ? OFFSET ?
	`, whereSQL)
	offset := (page - 1) * limit
	finalArgs := append(args, limit, offset)
	rows, err := database.DB.Query(selectQuery, finalArgs...)
	if err != nil {
		return models.ListEventsResponse{}, err
	}
	defer rows.Close()

	var events []models.Event
	var eventsList []models.ListEvent
	for rows.Next() {
		var e models.Event
		var metaBytes []byte
		err := rows.Scan(
			&e.EventID,
			&e.UserID,
			&e.EventType,
			&e.ObjectType,
			&e.ObjectID,
			&e.EventTime,
			&metaBytes,
		)
		if err != nil {
			log.Println("Scan row error: ", err)
			continue
		}
		json.Unmarshal(metaBytes, &e.Metadata)
		events = append(events, e)
		eventsList = append(eventsList, models.ListEvent{
			EventID:    e.EventID,
			UserID:     e.UserID,
			EventType:  e.EventType,
			ObjectType: util.SqlNullStringToPtr(e.ObjectType),
			ObjectID:   util.SqlNullInt64ToPtr(e.ObjectID),
			EventTime:  util.SqlNullTimeToPtr(e.EventTime),
			Metadata:   util.SqlNullMapToPtr(e.Metadata),
		})
	}
	if err := rows.Err(); err != nil {
		return models.ListEventsResponse{}, err
	}
	totalPages := int64(math.Ceil(float64(totalEvents) / float64(limit)))

	return models.ListEventsResponse{
		Events:      eventsList,
		Limit:       limit,
		Page:        page,
		TotalEvents: totalEvents,
		TotalPages:  totalPages,
	}, nil
}
