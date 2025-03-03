package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/auth"
	"VoizyServer/internal/util"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" && req.Username == "" {
		http.Error(w, "Missing required body params. Either 'email' or 'username' must be provided.", http.StatusBadRequest)
		return
	}

	response, err := login(req)
	if err != nil {
		http.Error(w, "Error logging in", http.StatusInternalServerError)
		return
	}

	if !response.IsPasswordCorrect {
		log.Println("INVALID PASSWORD ATTEMPTED for username: ", req.Username, ", email: ", req.Email)
		http.Error(w, "Email, Username, or Password is incorrect.", http.StatusUnauthorized)
		return
	}

	// Track successful Login event
	go util.TrackEvent(response.UserID, "login", "", nil, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func login(req models.LoginRequest) (models.LoginResponse, error) {
	var user models.User
	var query string
	var arg string

	if req.Email != "" {
		query = `
			SELECT user_id, email, username, password_hash, salt, api_key, created_at, updated_at
			FROM users
			WHERE email = ?
			LIMIT 1;
		`
		arg = req.Email
	} else {
		query = `
			SELECT user_id, email, username, password_hash, salt, api_key, created_at, updated_at
			FROM users
			WHERE username = ?
			LIMIT 1;
		`
		arg = req.Username
	}

	row := database.DB.QueryRow(query, arg)
	err := row.Scan(
		&user.UserID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Salt,
		&user.APIKey,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LoginResponse{}, fmt.Errorf("user not found: %w", err)
		}
		return models.LoginResponse{}, err
	}

	isPasswordCorrect := util.CheckPasswordHash(req.Password+user.Salt, user.PasswordHash)

	return models.LoginResponse{
		IsPasswordCorrect: isPasswordCorrect,
		UserID:            user.UserID,
		APIKey:            user.APIKey,
		Email:             user.Email,
		Username:          user.Username,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}, nil
}
