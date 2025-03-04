package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	if req.UserID <= 0 {
		http.Error(w, "Missing or invalid userID.", http.StatusBadRequest)
		return
	}

	response, err := createPost(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating post (%v).", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createPost(req models.CreatePostRequest) (models.CreatePostResponse, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		log.Println("Error beginning transaction: ", err)
		return models.CreatePostResponse{
			Success: false,
			Message: fmt.Sprintf("Error beginning transaction (%v).", err),
		}, fmt.Errorf("error beginning transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	postID, err := insertPost(tx, req)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to insert post: ", err)
		return models.CreatePostResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to insert post: %v", err),
		}, err
	}

	if req.IsPoll {
		err = insertPollOptions(tx, postID, req.PollOptions)
		if err != nil {
			tx.Rollback()
			log.Println("Failed to insert poll options: ", err)
			return models.CreatePostResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to insert poll options: %v", err),
			}, err
		}
	}

	err = insertPostMedia(tx, postID, req.Images)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to insert post media: ", err)
		return models.CreatePostResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to insert post media: %v", err),
		}, err
	}

	err = insertPostHashtags(tx, postID, req.Hashtags)
	if err != nil {
		tx.Rollback()
		log.Println("Failed to insert hashtags: ", err)
		return models.CreatePostResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to insert hashtags: %v", err),
		}, err
	}

	if err := tx.Commit(); err != nil {
		log.Println("Failed to commit transaction: ", err)
		return models.CreatePostResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to commit transaction: %v", err),
		}, err
	}

	return models.CreatePostResponse{
		Success: true,
		Message: "Post created successfully",
		PostID:  postID,
	}, nil
}

func insertPost(tx *sql.Tx, req models.CreatePostRequest) (int64, error) {
	if !req.IsPoll {
		query := `
			INSERT INTO posts (
				user_id,
				content_text,
				location_name,
				location_lat,
				location_lng,
				is_poll
			)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		result, err := tx.Exec(query,
			req.UserID,
			req.ContentText,
			req.LocationName,
			req.LocationLat,
			req.LocationLong,
			false,
		)
		if err != nil {
			log.Println("Error inserting into posts: ", err)
			return 0, err
		}
		return result.LastInsertId()
	}

	query := `
		INSERT INTO posts (
			user_id, content_text, location_name, location_lat, location_lng,
			is_poll, poll_question, poll_duration_type, poll_duration_length
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := tx.Exec(query,
		req.UserID,
		req.ContentText,
		req.LocationName,
		req.LocationLat,
		req.LocationLong,
		req.IsPoll,
		req.PollQuestion,
		req.PollDurationType,
		req.PollDurationLength,
	)
	if err != nil {
		log.Println("Error inserting into posts: ", err)
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error getting post lastInsertId: ", err)
		return 0, err
	}

	return postID, nil
}

func insertPollOptions(tx *sql.Tx, postID int64, options []string) error {
	if len(options) == 0 {
		log.Println("No options were passed to insertPollOptions...")
		return nil
	}

	query := `
		INSERT INTO poll_options (post_id, option_text)
		VALUES (?, ?)
	`
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println("Error preparing insert into poll_options: ", err)
		return err
	}
	defer stmt.Close()

	for _, opt := range options {
		if _, err := stmt.Exec(postID, opt); err != nil {
			log.Println("Error executing insert into poll_options: ", err)
			return err
		}
	}

	return nil
}

func insertPostMedia(tx *sql.Tx, postID int64, images []string) error {
	if len(images) == 0 {
		log.Println("No images were passed to insertPostMedia...")
		return nil
	}

	query := `
		INSERT INTO post_media (post_id, media_url, media_type)
		VALUES (?, ?, 'image')
	`
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println("Error preparing insert into post_media: ", err)
		return err
	}
	defer stmt.Close()

	for _, imgURL := range images {
		_, err = stmt.Exec(postID, imgURL)
		if err != nil {
			log.Println("Error executing insert into post_media: ", err)
			return err
		}
	}

	return nil
}

func insertPostHashtags(tx *sql.Tx, postID int64, tags []string) error {
	if len(tags) == 0 {
		log.Println("No tags were passed to insertPostHashtags...")
		return nil
	}

	upsertTag := `
		INSERT INTO hashtags (tag)
		VALUES (?)
		ON DUPLICATE KEY UPDATE tag=VALUES(tag)
	`
	insertTagStmt, err := tx.Prepare(upsertTag)
	if err != nil {
		log.Println("Error preparing upsertTag: ", err)
		return err
	}
	defer insertTagStmt.Close()

	selectTag := `
		SELECT hashtag_id FROM hashtags WHERE tag = ? LIMIT 1
	`
	selectTagStmt, err := tx.Prepare(selectTag)
	if err != nil {
		log.Println("Error preparing selectTag: ", err)
		return err
	}
	defer selectTagStmt.Close()

	insertPostHashtag := `
		INSERT INTO post_hashtags (post_id, hashtag_id)
		VALUES (?, ?)
	`
	postHashtagStmt, err := tx.Prepare(insertPostHashtag)
	if err != nil {
		log.Println("Error preparing insertPostHashtag: ", err)
		return err
	}
	defer postHashtagStmt.Close()

	for _, t := range tags {
		cleanedTag := t
		if len(cleanedTag) > 0 && cleanedTag[0] == '#' {
			cleanedTag = cleanedTag[1:]
		}

		_, err = insertTagStmt.Exec(cleanedTag)
		if err != nil {
			log.Println("Error executing upsertTag: ", err)
			return err
		}

		var tagID int64
		err = selectTagStmt.QueryRow(cleanedTag).Scan(&tagID)
		if err != nil {
			log.Println("Error executing selectTag: ", err)
			return err
		}

		_, err = postHashtagStmt.Exec(postID, tagID)
		if err != nil {
			log.Println("Error executing insertPostHashtag: ", err)
			return err
		}
	}

	return nil
}
