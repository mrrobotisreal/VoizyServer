package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
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

func listUserProfiles() (models.ListUsersResponse, error) {
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
		return models.ListUsersResponse{}, fmt.Errorf("error executing ListUserProfiles query: %w", err)
	}
	defer rows.Close()

	var profiles []models.UserProfile

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

		profiles = append(profiles, p)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error iterating over rows: ", err)
		return models.ListUsersResponse{}, fmt.Errorf("error iterating over rows: %w", err)
	}

	return models.ListUsersResponse{
		Profiles: profiles,
	}, nil
}
