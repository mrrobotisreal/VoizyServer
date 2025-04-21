package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"VoizyServer/internal/util"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func PutUserPreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.PutUserPreferencesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := putPreferences(req)
	if err != nil {
		log.Println("Failed to put user preferences due to the following error: ", err)
		http.Error(w, "Failed to put user preferences.", http.StatusInternalServerError)
		return
	}

	go func() {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}
		appVersion := r.Header.Get("X-App-Version")
		if appVersion == "" {
			appVersion = "v0.0.1"
		}
		osVersion := r.Header.Get("X-OS-Version")
		if osVersion == "" {
			osVersion = "Unknown"
		}
		deviceModel := r.Header.Get("X-Device-Model")
		if deviceModel == "" {
			deviceModel = "Unknown"
		}
		metadata := map[string]interface{}{
			"client_ip":    ip,
			"user_agent":   r.Header.Get("User-Agent"),
			"app_version":  appVersion,
			"os_version":   osVersion,
			"device_model": deviceModel,
		}
		util.TrackEvent(req.UserID, "put_user_preferences", "user_preferences", &response.UserPreferencesID, metadata)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func putPreferences(req models.PutUserPreferencesRequest) (models.PutUserPreferencesResponse, error) {
	var prefsID int64
	err := database.DB.
		QueryRow(`SELECT user_preferences_id 
                  FROM user_preferences 
                  WHERE user_id = ?`, req.UserID).
		Scan(&prefsID)
	if err != nil && err != sql.ErrNoRows {
		return models.PutUserPreferencesResponse{
			Success: false,
			Message: fmt.Sprintf("error checking existing preferences: %v", err),
		}, err
	}

	cols := []string{"user_id"}
	vals := []interface{}{req.UserID}
	placeholders := []string{"?"}

	if req.PrimaryColor != nil {
		cols = append(cols, "primary_color")
		placeholders = append(placeholders, "?")
		vals = append(vals, *req.PrimaryColor)
	}
	if req.PrimaryAccent != nil {
		cols = append(cols, "primary_accent")
		placeholders = append(placeholders, "?")
		vals = append(vals, *req.PrimaryAccent)
	}
	if req.SecondaryColor != nil {
		cols = append(cols, "secondary_color")
		placeholders = append(placeholders, "?")
		vals = append(vals, *req.SecondaryColor)
	}
	if req.SecondaryAccent != nil {
		cols = append(cols, "secondary_accent")
		placeholders = append(placeholders, "?")
		vals = append(vals, *req.SecondaryAccent)
	}
	if req.SongAutoplay != nil {
		cols = append(cols, "song_autoplay")
		placeholders = append(placeholders, "?")
		vals = append(vals, *req.SongAutoplay)
	}
	if req.ProfilePrimaryColor != nil {
		cols = append(cols, "profile_primary_color")
		placeholders = append(placeholders, "?")
		vals = append(vals, *req.ProfilePrimaryColor)
	}
	if req.ProfilePrimaryAccent != nil {
		cols = append(cols, "profile_primary_accent")
		placeholders = append(placeholders, "?")
		vals = append(vals, *req.ProfilePrimaryAccent)
	}
	if req.ProfileSecondaryColor != nil {
		cols = append(cols, "profile_secondary_color")
		placeholders = append(placeholders, "?")
		vals = append(vals, *req.ProfileSecondaryColor)
	}
	if req.ProfileSecondaryAccent != nil {
		cols = append(cols, "profile_secondary_accent")
		placeholders = append(placeholders, "?")
		vals = append(vals, *req.ProfileSecondaryAccent)
	}
	if req.ProfileSongAutoplay != nil {
		cols = append(cols, "profile_song_autoplay")
		placeholders = append(placeholders, "?")
		vals = append(vals, *req.ProfileSongAutoplay)
	}

	if err == sql.ErrNoRows {
		query := fmt.Sprintf(
			"INSERT INTO user_preferences (%s) VALUES (%s)",
			strings.Join(cols, ", "),
			strings.Join(placeholders, ", "),
		)
		res, err := database.DB.Exec(query, vals...)
		if err != nil {
			return models.PutUserPreferencesResponse{
				Success: false,
				Message: fmt.Sprintf("error inserting preferences: %v", err),
			}, err
		}
		newID, _ := res.LastInsertId()
		return models.PutUserPreferencesResponse{
			Success:           true,
			Message:           "Preferences created",
			UserPreferencesID: newID,
		}, nil
	}

	setClauses := make([]string, 0, len(cols)-1)
	updateVals := make([]interface{}, 0, len(cols)-1)
	for i, col := range cols {
		if col == "user_id" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
		updateVals = append(updateVals, vals[i])
	}
	updateVals = append(updateVals, prefsID)

	updQuery := fmt.Sprintf(
		"UPDATE user_preferences SET %s WHERE user_preferences_id = ?",
		strings.Join(setClauses, ", "),
	)
	_, err = database.DB.Exec(updQuery, updateVals...)
	if err != nil {
		return models.PutUserPreferencesResponse{
			Success: false,
			Message: fmt.Sprintf("error updating preferences: %v", err),
		}, err
	}
	return models.PutUserPreferencesResponse{
		Success:           true,
		Message:           "Preferences updated",
		UserPreferencesID: prefsID,
	}, nil
}
