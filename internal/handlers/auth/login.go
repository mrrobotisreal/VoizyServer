package handlers

import (
	"VoizyServer/internal/database"
	"VoizyServer/internal/database/firebase"
	models "VoizyServer/internal/models/auth"
	"VoizyServer/internal/util"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	go util.TrackEvent(response.UserID, "login", "user", &response.UserID, map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func login(req models.LoginRequest) (models.LoginResponse, error) {
	ctx := context.Background()

	email := req.Email
	if email == "" {
		email = lookupEmailByUsername(req.Username)
	}

	signIn, err := firebase.SignInWithEmail(ctx, email, req.Password)
	if err != nil {
		return models.LoginResponse{}, err
	}

	_, err = firebase.AuthClient.VerifyIDToken(ctx, signIn.IDToken)
	if err != nil {
		return models.LoginResponse{}, err
	}

	var user models.User
	var query string
	var arg string

	if email != "" {
		query = `
			SELECT user_id, fb_uid, email, phone, username, password_hash, salt, api_key, created_at, updated_at
			FROM users
			WHERE email = ?
			LIMIT 1;
		`
		arg = email
	} else {
		query = `
			SELECT user_id, fb_uid, email, phone, username, password_hash, salt, api_key, created_at, updated_at
			FROM users
			WHERE username = ?
			LIMIT 1;
		`
		arg = req.Username
	}

	var fbuid sql.NullString
	var phone sql.NullString
	row := database.DB.QueryRow(query, arg)
	err = row.Scan(
		&user.UserID,
		&fbuid,
		&user.Email,
		&phone,
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
	user.FBUID = util.SqlNullStringToPtr(fbuid)
	user.Phone = util.SqlNullStringToPtr(phone)

	log.Println("What is userID? ", user.UserID)
	isPasswordCorrect := util.CheckPasswordHash(req.Password+user.Salt, user.PasswordHash)
	token, err := util.GenerateAndStoreJWT(strconv.FormatInt(user.UserID, 10), "always") // TODO: implement sessionOptions
	if err != nil {
		log.Println("Failed to generate JWT: ", err)
		return models.LoginResponse{}, err
	}

	return models.LoginResponse{
		IsPasswordCorrect: isPasswordCorrect,
		UserID:            user.UserID,
		FBUID:             user.FBUID,
		APIKey:            user.APIKey,
		Token:             token,
		Email:             user.Email,
		Phone:             user.Phone,
		Username:          user.Username,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}, nil
}

func lookupEmailByUsername(username string) string {
	var email string
	query := `
		SELECT email FROM users WHERE username = ?;
	`
	_ = database.DB.QueryRow(query, username).Scan(&email)

	return email
}
