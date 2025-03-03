package util

import (
	"VoizyServer/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func TrackEvent(userID int64, eventType, objectType string, objectID *int64, metadata map[string]interface{}) error {
	var metaBytes []byte
	var err error
	if metadata != nil {
		metaBytes, err = json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	var result sql.Result
	if objectID != nil {
		query := `
			INSERT INTO analytics_events (user_id, event_type, object_type, object_id, event_time, metadata)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		result, err = database.DB.Exec(query, userID, eventType, objectType, *objectID, time.Now(), metaBytes)
	} else {
		query := `
			INSERT INTO analytics_events (user_id, event_type, object_type, event_time, metadata)
			VALUES (?, ?, ?, ?, ?)
		`
		result, err = database.DB.Exec(query, userID, eventType, objectType, time.Now(), metaBytes)
	}

	if err != nil {
		return fmt.Errorf("TrackEvent DB insert error: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Println("TrackEvent success; total rows affected = ", rowsAffected)
	} else {
		log.Println("TrackEvent fail; no rows affected - rows affected = ", rowsAffected)
	}

	return nil
}
