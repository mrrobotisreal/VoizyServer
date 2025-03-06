package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/analytics"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ListStatsHandler(w http.ResponseWriter, r *http.Request) {
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

	groupBy := q.Get("group_by")
	if groupBy == "" {
		groupBy = "day"
	}

	response, err := listStats(userID, eventType, objectType, startTime, endTime, groupBy)
	if err != nil {
		log.Println("Failed to list events due to the following error: ", err)
		http.Error(w, "Failed to list events.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listStats(userID int64, eventType, objectType string, startTime, endTime time.Time, groupBy string) (models.ListStatsResponse, error) {
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

	var groupExpr string
	switch groupBy {
	case "day":
		groupExpr = "DATE_FORMAT(event_time, '%Y-%m-%d')"
	case "hour":
		groupExpr = "DATE_FORMAT(event_time, '%Y-%m-%d %H:00')"
	case "month":
		groupExpr = "DATE_FORMAT(event_time, '%Y-%m')"
	case "event_type":
		groupExpr = "event_type"
	case "object_type":
		groupExpr = "object_type"
	default:
		groupExpr = "DATE_FORMAT(event_time, '%Y-%m-%d')"
	}

	query := fmt.Sprintf(`
		SELECT
			%s AS group_value,
			COUNT(*) AS count
		FROM analytics_events
		WHERE %s
		GROUP BY group_value
		ORDER BY group_value
	`, groupExpr, whereSQL)
	log.Println("Stats query: ", query, args)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return models.ListStatsResponse{}, err
	}
	defer rows.Close()

	var stats []models.StatSummary
	for rows.Next() {
		var s models.StatSummary
		if err := rows.Scan(&s.GroupValue, &s.Count); err != nil {
			log.Println("Scan row error: ", err)
			continue
		}
		stats = append(stats, models.StatSummary{
			GroupLabel: groupBy,
			GroupValue: s.GroupValue,
			Count:      s.Count,
		})
	}

	if err := rows.Err(); err != nil {
		return models.ListStatsResponse{}, err
	}

	return models.ListStatsResponse{
		Stats: stats,
	}, nil
}
