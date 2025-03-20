package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/users"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
)

func ListSongsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	userIDString := q.Get("id")
	if userIDString == "" {
		http.Error(w, "Missing required param 'id'.", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		log.Println("Failed to parse userIDString (string) to userID (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'id'.", http.StatusInternalServerError)
		return
	}

	limitString := q.Get("limit")
	if limitString == "" {
		http.Error(w, "Missing required param 'limit'.", http.StatusBadRequest)
		return
	}
	limit, err := strconv.ParseInt(limitString, 10, 64)
	if err != nil {
		log.Println("Failed to parse limitString (string) to limit (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'limit'.", http.StatusInternalServerError)
		return
	}

	pageString := q.Get("page")
	if pageString == "" {
		http.Error(w, "Missing required param 'page'.", http.StatusBadRequest)
		return
	}
	page, err := strconv.ParseInt(pageString, 10, 64)
	if err != nil {
		log.Println("Failed to parse pageString (string) to page (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'page'.", http.StatusInternalServerError)
		return
	}

	response, err := listSongs(limit, page)
	if err != nil {
		log.Println("Failed to list songs due to the following error: ", err)
		http.Error(w, "Failed to list songs.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listSongs(limit, page int64) (models.ListSongsResponse, error) {
	offset := (page - 1) * limit

	var totalSongs int64
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM songs").Scan(&totalSongs); err != nil {
		return models.ListSongsResponse{}, err
	}

	rows, err := database.DB.Query("SELECT * FROM songs LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return models.ListSongsResponse{}, err
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var song models.Song
		if err := rows.Scan(&song.SongID, &song.Title, &song.Artist, &song.SongURL); err != nil {
			continue
		}
		songs = append(songs, song)
	}

	totalPages := int64(math.Ceil(float64(totalSongs) / float64(limit)))

	return models.ListSongsResponse{
		Songs:      songs,
		Limit:      limit,
		Page:       page,
		TotalSongs: totalSongs,
		TotalPages: totalPages,
	}, nil
}
