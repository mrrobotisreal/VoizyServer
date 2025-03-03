package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func UpdateUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	profileIDString := r.URL.Query().Get("id")
	if profileIDString == "" {
		http.Error(w, "Missing required param 'id'.", http.StatusBadRequest)
		return
	}
	profileID, err := strconv.ParseInt(profileIDString, 10, 64)
	if err != nil {
		http.Error(w, "Error parsing 'id'.", http.StatusInternalServerError)
		return
	}

	var req map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := updateUserProfile(profileID, req)
	if err != nil {
		http.Error(w, "Error updating user profile.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateUserProfile(profileID int64, req map[string]interface{}) (models.UpdateUserProfileResponse, error) {
	validColumns := map[string]string{
		"firstName":       "first_name",
		"lastName":        "last_name",
		"preferredName":   "preferred_name",
		"birthDate":       "birth_date",
		"cityOfResidence": "city_of_residence",
		"placeOfWork":     "place_of_work",
	}
	setClauses := []string{}
	args := []interface{}{}

	for key, val := range req {
		colName, ok := validColumns[key]
		if !ok {
			continue
		}
		if val == nil {
			setClauses = append(setClauses, fmt.Sprintf("%s = NULL", colName))
		} else if key == "birth_date" && val != nil {
			strVal, ok := val.(string)
			if !ok {
				log.Println("birth_date must be a string in YYYY-MM-DD format")
				return models.UpdateUserProfileResponse{
					IsUpdateSuccessful: false,
				}, fmt.Errorf("birth_date must be a string in YYYY-MM-DD format")
			}
			t, err := time.Parse("2025-03-02", strVal)
			if err != nil {
				log.Println("Invalid date format (YYYY-MM-DD)")
				return models.UpdateUserProfileResponse{
					IsUpdateSuccessful: false,
				}, fmt.Errorf("invalid date format (YYYY-MM-DD)")
			}
			setClauses = append(setClauses, "birth_date = ?")
			val = t.Format("2025-03-02")
			args = append(args, val)
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", colName))
			args = append(args, val)
		}
	}

	if len(setClauses) == 0 {
		log.Println("No updatable fields provided.")
		return models.UpdateUserProfileResponse{
			IsUpdateSuccessful: false,
		}, fmt.Errorf("no updatable fields provided")
	}

	query := fmt.Sprintf(
		"UPDATE user_profiles SET %s WHERE profile_id = ?",
		joinClauses(setClauses, ", "))
	args = append(args, profileID)

	result, err := database.DB.Exec(query, args...)
	if err != nil {
		log.Println("UpdateUserProfile - DB error: ", err)
		return models.UpdateUserProfileResponse{
			IsUpdateSuccessful: false,
		}, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Println("No rows updated (invalid profile_id?)")
		return models.UpdateUserProfileResponse{
			IsUpdateSuccessful: false,
		}, fmt.Errorf("no rows updated")
	}

	return models.UpdateUserProfileResponse{
		IsUpdateSuccessful: true,
	}, nil
}

func joinClauses(clauses []string, sep string) string {
	out := ""
	for i, c := range clauses {
		if i > 0 {
			out += sep
		}
		out += c
	}
	return out
}
