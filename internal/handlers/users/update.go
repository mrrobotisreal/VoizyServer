package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"VoizyServer/internal/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userIDString := r.URL.Query().Get("id")
	if userIDString == "" {
		http.Error(w, "Missing required param 'id'.", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		http.Error(w, "Error parsing 'id'.", http.StatusInternalServerError)
		return
	}

	var req models.UpdateUserRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := updateUser(userID, req)
	if err != nil {
		http.Error(w, "Error updating user.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(userID, "update_user", "user", &userID, map[string]interface{}{
		"username": req.Username,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateUser(userID int64, req models.UpdateUserRequest) (models.UpdateUserResponse, error) {
	query := `
		UPDATE users
		SET username = ?
		WHERE user_id = ?
	`

	result, err := database.DB.Exec(query, req.Username, userID)
	if err != nil {
		log.Println("UpdateUser - DB error: ", err)
		return models.UpdateUserResponse{
			IsUpdateSuccessful: false,
		}, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Println("No user updated (invalid user_id?)")
		return models.UpdateUserResponse{
			IsUpdateSuccessful: false,
		}, fmt.Errorf("no user updated; 0 rows affected; for user_id %d\n", userID)
	}

	return models.UpdateUserResponse{
		IsUpdateSuccessful: true,
	}, nil
}
