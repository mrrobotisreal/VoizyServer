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

func ListImagesHandler(w http.ResponseWriter, r *http.Request) {
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

	response, err := listImages(userID, limit, page)
	if err != nil {
		log.Println("Failed to list images due to the following reason: ", err)
		http.Error(w, "Failed to list images.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listImages(userID, limit, page int64) (models.ListImagesResponse, error) {
	var totalImages int64
	countQuery := `
		SELECT COUNT(*)
		FROM user_images
		WHERE user_id = ?
	`
	err := database.DB.QueryRow(countQuery, userID).Scan(&totalImages)
	if err != nil {
		return models.ListImagesResponse{}, err
	}

	offset := (page - 1) * limit
	selectQuery := `
		SELECT
			user_id,
			user_image_id,
			image_url,
			is_profile_pic,
			uploaded_at
		FROM user_images
		WHERE user_id = ?
		ORDER BY uploaded_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := database.DB.Query(selectQuery, userID, limit, offset)
	if err != nil {
		return models.ListImagesResponse{}, err
	}
	defer rows.Close()

	var images []models.UserImage
	for rows.Next() {
		var i models.UserImage
		err := rows.Scan(
			&i.UserID,
			&i.UserImageID,
			&i.ImageURL,
			&i.IsProfilePicture,
			&i.UploadedAt,
		)
		if err != nil {
			log.Println("Scan row error: ", err)
			continue
		}
		images = append(images, i)
	}
	if err := rows.Err(); err != nil {
		return models.ListImagesResponse{}, err
	}
	totalPages := int64(math.Ceil(float64(totalImages) / float64(limit)))

	return models.ListImagesResponse{
		Images:      images,
		Limit:       limit,
		Page:        page,
		TotalImages: totalImages,
		TotalPages:  totalPages,
	}, nil
}
