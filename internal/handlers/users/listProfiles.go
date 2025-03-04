package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"VoizyServer/internal/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func ListUserProfilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	response, err := listUserProfiles()
	if err != nil {
		http.Error(w, "Error listing users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listUserProfiles() (models.ListProfilesResponse, error) {
	query := `
		SELECT
			profile_id,
			user_id,
			IFNULL(first_name, ''),
			IFNULL(last_name, ''),
			IFNULL(preferred_name, ''),
			birth_date,
			IFNULL(city_of_residence, ''),
			IFNULL(place_of_work, ''),
			date_joined
		FROM user_profiles;
	`
	rows, err := database.DB.Query(query)
	if err != nil {
		log.Println("Error executing ListUserProfiles query: ", err)
		return models.ListProfilesResponse{}, fmt.Errorf("error executing ListUserProfiles query: %w", err)
	}
	defer rows.Close()

	var profiles []models.ListProfile

	for rows.Next() {
		var p models.UserProfile

		err := rows.Scan(
			&p.ProfileID,
			&p.UserID,
			&p.FirstName,
			&p.LastName,
			&p.PreferredName,
			&p.BirthDate,
			&p.CityOfResidence,
			&p.PlaceOfWork,
			&p.DateJoined,
		)
		if err != nil {
			log.Println("Scan error: ", err)
			continue
		}

		profiles = append(profiles, models.ListProfile{
			UserID:          p.UserID,
			ProfileID:       p.ProfileID,
			FirstName:       util.SqlNullStringToPtr(p.FirstName),
			LastName:        util.SqlNullStringToPtr(p.LastName),
			PreferredName:   util.SqlNullStringToPtr(p.PreferredName),
			BirthDate:       util.SqlNullTimeToPtr(p.BirthDate),
			CityOfResidence: util.SqlNullStringToPtr(p.CityOfResidence),
			PlaceOfWork:     util.SqlNullStringToPtr(p.PlaceOfWork),
			DateJoined:      util.SqlNullTimeToPtr(p.DateJoined),
		})
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating over rows: ", err)
		return models.ListProfilesResponse{}, fmt.Errorf("error iterating over rows: %w", err)
	}

	return models.ListProfilesResponse{
		Profiles: profiles,
	}, nil
}
