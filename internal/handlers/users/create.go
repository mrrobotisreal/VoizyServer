package handlers

import (
	"VoizyServer/internal/database"
	"VoizyServer/internal/database/firebase"
	models "VoizyServer/internal/models/users"
	"VoizyServer/internal/util"
	"context"
	"firebase.google.com/go/v4/auth"
	"fmt"

	//"context"
	"encoding/json"
	//"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := createUser(req)
	if err != nil {
		http.Error(w, "Error creating the user", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(response.UserID, "create_account", "user", &response.UserID, map[string]interface{}{
		"email":    response.Email,
		"username": response.Username,
	})
	go util.TrackEvent(response.UserID, "create_profile", "user_profile", &response.ProfileID, map[string]interface{}{
		"preferredName": response.PreferredName,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createUser(req models.CreateUserRequest) (models.CreateUserResponse, error) {
	ctx := context.Background()

	params := (&auth.UserToCreate{}).Email(req.Email).Password(req.Password).DisplayName(req.PreferredName)

	u, err := firebase.AuthClient.CreateUser(ctx, params)
	if err != nil {
		return models.CreateUserResponse{}, err
	}
	fmt.Println("Firebase UID = ", u.UID)

	apiKey, err := util.GenerateSecureAPIKey()
	if err != nil {
		log.Println("Error generating API key: ", err)
		return models.CreateUserResponse{}, err
	}
	//fbToken, _ := util.GenerateAndStoreJWT(u.UID, "always")

	// TODO: update this to use Tx and do rollbacks upon any failures
	//salt, err := util.GenerateSalt(10)
	//if err != nil {
	//	log.Println("Error generating salt: ", err)
	//	return models.CreateUserResponse{}, err
	//}

	//hashedPassword, err := util.HashPassword(req.Password + salt)
	//if err != nil {
	//	log.Println("Error hashing password: ", err)
	//	return models.CreateUserResponse{}, err
	//}

	currentTime := time.Now().UTC()
	userQuery := `INSERT INTO users (fb_uid, email, api_key, salt, password_hash, username, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	userResult, err := database.DB.Exec(userQuery, u.UID, req.Email, apiKey.Key, "", "", req.Username, currentTime, currentTime)
	if err != nil {
		log.Println("CreateUserHandler - DB error: ", err)
		return models.CreateUserResponse{}, err
	}
	userID, _ := userResult.LastInsertId()

	now := currentTime
	expiresAt := now.AddDate(0, 0, 90)
	apiKeyQuery := `
		INSERT INTO api_keys (
			user_id,
			api_key,
			created_at,
			expires_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err = database.DB.Exec(apiKeyQuery, userID, apiKey.Key, currentTime, expiresAt, currentTime)
	if err != nil {
		log.Println("InsertAPIKey - DB error: ", err) // TODO: update all these logs to follow format of rest of package
		return models.CreateUserResponse{}, err
	}

	token, err := util.GenerateAndStoreJWT(strconv.FormatInt(userID, 10), "always") // TODO: implement sessionOptions
	if err != nil {
		log.Println("GenerateJWT error: ", err)
		return models.CreateUserResponse{}, err
	}

	profileQuery := `INSERT INTO user_profiles (user_id, first_name, last_name, preferred_name, birth_date, city_of_residence, place_of_work, date_joined) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	profileResult, err := database.DB.Exec(profileQuery, userID, req.PreferredName, nil, req.PreferredName, nil, nil, nil, currentTime)
	if err != nil {
		log.Println("CreateUserProfile - DB error: ", err)
		return models.CreateUserResponse{}, err
	}
	profileID, _ := profileResult.LastInsertId()

	//ctx := context.Background()
	//key := fmt.Sprintf("user:%d:username", userID)
	//err = database.RDB.Set(ctx, key, req.Username, 0).Err()
	//if err != nil {
	//	log.Println("Redis set error: ", err)
	//}

	return models.CreateUserResponse{
		UserID:        userID,
		FBUID:         u.UID,
		ProfileID:     profileID,
		APIKey:        apiKey.Key,
		Token:         token,
		Email:         req.Email,
		Phone:         "",
		Username:      req.Username,
		PreferredName: req.PreferredName,
		FirstName:     req.PreferredName,
		DateJoined:    currentTime,
		CreatedAt:     currentTime,
		UpdatedAt:     currentTime,
	}, nil
}
