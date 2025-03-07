package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/auth"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func InsertApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.InsertApiKeyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := insertApiKey(req)
	if err != nil {
		log.Println("Failed to insert API Key due to the following error: ", err)
		http.Error(w, "Failed to insert API Key.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func insertApiKey(req models.InsertApiKeyRequest) (models.InsertApiKeyResponse, error) {
	now := time.Now().UTC()
	timeLayout := "2006-01-02 15:04:05"
	createdAt := now.Format(timeLayout)
	expiresAt := now.AddDate(0, 0, 90).Format(timeLayout)
	updatedAt := now.Format(timeLayout)
	query := `
		INSERT INTO api_keys (
			user_id,
			api_key,
			created_at,
			expires_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := database.DB.Exec(query, req.UserID, req.APIKey, createdAt, expiresAt, updatedAt)
	if err != nil {
		return models.InsertApiKeyResponse{
			Success: false,
		}, err
	}
	apiKeyID, _ := result.LastInsertId()

	return models.InsertApiKeyResponse{
		Success:  true,
		APIKeyID: apiKeyID,
	}, nil
}
