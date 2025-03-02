package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	email := r.URL.Query().Get("email")
	if username == "" && email == "" {
		http.Error(w, "Missing required params. You must provide either 'username' or 'email'.", http.StatusBadRequest)
		return
	}

	response, err := getUser(username, email)
	if err != nil {
		http.Error(w, "Error getting the user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getUser(username, email string) (models.GetUserResponse, error) {
	var response models.GetUserResponse
	var query string
	var arg string

	if username != "" {
		query = `
			SELECT user_id, email, username, created_at, updated_at
			FROM users
			WHERE username = ?
			LIMIT 1
		`
		arg = username
	} else {
		query = `
			SELECT user_id, email, username, created_at, updated_at
			FROM users
			WHERE email = ?
			LIMIT 1
		`
		arg = email
	}

	row := database.DB.QueryRow(query, arg)
	err := row.Scan(
		&response.UserID,
		&response.Email,
		&response.Username,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.GetUserResponse{}, fmt.Errorf("user not found: %w", err)
		}
		return models.GetUserResponse{}, err
	}

	return response, nil
}
