package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func GetPostMediaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	postIDString := r.URL.Query().Get("id")
	if postIDString == "" {
		http.Error(w, "Missing required param 'id'.", http.StatusBadRequest)
		return
	}
	postID, err := strconv.ParseInt(postIDString, 10, 64)
	if err != nil {
		log.Println("Failed to convert postIDString (string) to postID (int64) due to the following error: ", err)
		http.Error(w, "Failed to convert param 'id'.", http.StatusInternalServerError)
		return
	}

	response, err := getPostMedia(postID)
	if err != nil {
		log.Println("Failed to get post media due to the following error: ", err)
		http.Error(w, "Failed to get post media.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getPostMedia(postID int64) (models.GetMediaResponse, error) {
	var images []string
	queryImages := `
		SELECT media_url
		FROM post_media
		WHERE post_id = ?
		AND media_type = 'image'
	`
	rows, err := database.DB.Query(queryImages, postID)
	if err != nil {
		return models.GetMediaResponse{}, err
	}
	for rows.Next() {
		var i string
		err := rows.Scan(&i)
		if err != nil {
			log.Println("Scan rows error: ", err)
			continue
		}
		images = append(images, i)
	}
	if err = rows.Err(); err != nil {
		return models.GetMediaResponse{}, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	var videos []string
	queryVideos := `
		SELECT media_url
		FROM post_media
		WHERE post_id = ?
		AND media_type = 'video'
	`
	rows, err = database.DB.Query(queryVideos, postID)
	if err != nil {
		return models.GetMediaResponse{}, err
	}
	for rows.Next() {
		var v string
		err := rows.Scan(&v)
		if err != nil {
			log.Println("Scan rows error: ", err)
			continue
		}
		videos = append(videos, v)
	}
	if err = rows.Err(); err != nil {
		return models.GetMediaResponse{}, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	// Returning an empty array for videos for now, as I have not implemented that aspect yet and there won't be any videos
	videos = []string{}
	return models.GetMediaResponse{
		Images: images,
		Videos: videos,
	}, nil
}
