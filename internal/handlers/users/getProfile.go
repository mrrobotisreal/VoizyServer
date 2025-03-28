package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"VoizyServer/internal/util"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userIDString := r.URL.Query().Get("id")
	if userIDString == "" {
		fmt.Println("'id' param is missing for getProfile")
		http.Error(w, "Missing required param 'id'.", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		fmt.Println("An error occurred while trying to parse 'id' into an int64")
		http.Error(w, "Error parsing 'id'", http.StatusInternalServerError)
		return
	}

	response, err := getProfile(userID)
	if err != nil {
		fmt.Println("An error occurred while trying to get the profile from the database: ", err)
		http.Error(w, "Error getting user profile.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(userID, "view_profile", "user_profile", &response.ProfileID, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getProfile(userID int64) (models.GetUserProfileResponse, error) {
	var profile models.UserProfile

	query := `
		SELECT p.profile_id, p.user_id, p.first_name, p.last_name, p.preferred_name, p.birth_date, p.city_of_residence, p.place_of_work, p.date_joined, u.username
		FROM user_profiles p
		LEFT JOIN users u ON p.user_id = u.user_id
		WHERE p.user_id = ?
		LIMIT 1
	`

	row := database.DB.QueryRow(query, userID)
	err := row.Scan(
		&profile.ProfileID,
		&profile.UserID,
		&profile.FirstName,
		&profile.LastName,
		&profile.PreferredName,
		&profile.BirthDate,
		&profile.CityOfResidence,
		&profile.PlaceOfWork,
		&profile.DateJoined,
		&profile.Username,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.GetUserProfileResponse{}, fmt.Errorf("user profile not found: %w", err)
		}
		return models.GetUserProfileResponse{}, err
	}

	return models.GetUserProfileResponse{
		ProfileID:       profile.ProfileID,
		UserID:          profile.UserID,
		FirstName:       util.SqlNullStringToPtr(profile.FirstName),
		LastName:        util.SqlNullStringToPtr(profile.LastName),
		PreferredName:   util.SqlNullStringToPtr(profile.PreferredName),
		BirthDate:       util.SqlNullTimeToPtr(profile.BirthDate),
		CityOfResidence: util.SqlNullStringToPtr(profile.CityOfResidence),
		PlaceOfWork:     util.SqlNullStringToPtr(profile.PlaceOfWork),
		DateJoined:      util.SqlNullTimeToPtr(profile.DateJoined),
		Username:        util.SqlNullStringToPtr(profile.Username),
	}, nil
}
