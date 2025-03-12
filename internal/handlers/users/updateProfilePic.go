package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func UpdateProfilePicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.UpdateProfilePicRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := updateProfilePic(req)
	if err != nil {
		log.Print("Failed to update profile pic due to the following error: ", err)
		http.Error(w, "Failed to update profile pic.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateProfilePic(req models.UpdateProfilePicRequest) (models.UpdateProfilePicResponse, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return models.UpdateProfilePicResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update profile pic due to: %v", err),
		}, err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	resetQuery := `
		UPDATE user_images
		SET is_profile_pic = 0
		WHERE user_id = ?
	`
	if _, err := tx.Exec(resetQuery, req.UserID); err != nil {
		tx.Rollback()
		return models.UpdateProfilePicResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update profile pic due to: %v", err),
		}, err
	}

	setQuery := `
		UPDATE user_images
		SET is_profile_pic = 1
		WHERE user_id = ?
			AND user_image_id = ?
	`
	result, err := tx.Exec(setQuery, req.UserID, req.ImageID)
	if err != nil {
		tx.Rollback()
		return models.UpdateProfilePicResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update profile pic due to: %v", err),
		}, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		return models.UpdateProfilePicResponse{
			Success: false,
			Message: fmt.Sprintf("No rows affected!"),
		}, fmt.Errorf("no rows affected")
	}

	if err := tx.Commit(); err != nil {
		return models.UpdateProfilePicResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update profile pic due to: %v", err),
		}, err
	}

	return models.UpdateProfilePicResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully updated profile pic!"),
	}, nil
}
