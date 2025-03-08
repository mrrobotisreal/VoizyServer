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

func CreateFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateFriendRequestRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := createFriendRequest(req)
	if err != nil {
		log.Println("Failed to create friend request due to the following error: ", err)
		http.Error(w, "Failed to create friend request.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(req.UserID, "create_friend_request", "friendship", &response.FriendshipID, map[string]interface{}{
		"friendID": req.FriendID,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createFriendRequest(req models.CreateFriendRequestRequest) (models.CreateFriendRequestResponse, error) {
	query := `
		INSERT INTO friendships (
			user_id,
			friend_id
		) VALUES (?, ?)
	`
	result, err := database.DB.Exec(query, req.UserID, req.FriendID)
	if err != nil {
		return models.CreateFriendRequestResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create friend request due to the following error: %v", err),
		}, err
	}
	friendshipID, _ := result.LastInsertId()

	return models.CreateFriendRequestResponse{
		Success:      true,
		Message:      "Successfully created friend request.",
		FriendshipID: friendshipID,
	}, nil
}
