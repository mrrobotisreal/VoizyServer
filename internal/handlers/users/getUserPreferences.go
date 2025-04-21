package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userIDString := r.URL.Query().Get("id")
	if userIDString == "" {
		http.Error(w, "Missing required param 'userID'.", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		log.Println("Error converting userIDString (string) to userID (int64): ", err)
		http.Error(w, "Failed to parse param 'id'. It should be an int >= 1.", http.StatusInternalServerError)
		return
	}

	response, err := getPreferences(userID)
	if err != nil {
		log.Println("Failed to get user preferences with the following error: ", err)
		http.Error(w, "failed to get user preferences.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getPreferences(userID int64) (models.GetUserPreferencesResponse, error) {
	var response models.GetUserPreferencesResponse

	var prefsID int64
	err := database.DB.
		QueryRow(`SELECT user_preferences_id
                  FROM user_preferences
                  WHERE user_id = ?`, userID).
		Scan(&prefsID)

	if err != nil && err != sql.ErrNoRows {
		return response, fmt.Errorf("error checking existing preferences: %w", err)
	}

	if err == sql.ErrNoRows {
		if _, err := database.DB.Exec(
			`INSERT INTO user_preferences (user_id) VALUES (?)`,
			userID,
		); err != nil {
			return response, fmt.Errorf("error creating default preferences: %w", err)
		}
	}

	const selectQuery = `
      SELECT
        user_id,
        primary_color,
        primary_accent,
        secondary_color,
        secondary_accent,
        song_autoplay,
        profile_primary_color,
        profile_primary_accent,
        profile_secondary_color,
        profile_secondary_accent,
        profile_song_autoplay
      FROM user_preferences
      WHERE user_id = ?
      LIMIT 1
    `
	err = database.DB.
		QueryRow(selectQuery, userID).
		Scan(
			&response.UserID,
			&response.PrimaryColor,
			&response.PrimaryAccent,
			&response.SecondaryColor,
			&response.SecondaryAccent,
			&response.SongAutoplay,
			&response.ProfilePrimaryColor,
			&response.ProfilePrimaryAccent,
			&response.ProfileSecondaryColor,
			&response.ProfileSecondaryAccent,
			&response.ProfileSongAutoplay,
		)
	if err != nil {
		return response, fmt.Errorf("error fetching preferences: %w", err)
	}

	return response, nil
}
