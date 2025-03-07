package handlers

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/posts"
	"VoizyServer/internal/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	postIDString := q.Get("post_id")
	if postIDString == "" {
		http.Error(w, "Missing required param 'post_id'.", http.StatusBadRequest)
		return
	}
	postID, err := strconv.ParseInt(postIDString, 10, 64)
	if err != nil {
		log.Println("Failed to parse postIDString (string) to postID (int64) due to the following error: ", err)
		http.Error(w, "Failed to parse param 'post_id'.", http.StatusInternalServerError)
		return
	}
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

	var req map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := updatePost(postID, userID, req)
	if err != nil {
		log.Println("Failed to update post due to the following error: ", err)
		http.Error(w, "Failed to update post.", http.StatusInternalServerError)
		return
	}

	go util.TrackEvent(userID, "update_post", "post", &postID, req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updatePost(postID, userID int64, req map[string]interface{}) (models.UpdatePostResponse, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return models.UpdatePostResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to start transaction - %v", err),
		}, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	setClauses := []string{}
	args := []interface{}{}

	handleStringOrNull(&setClauses, &args, req, "contentText", "content_text")
	handleStringOrNull(&setClauses, &args, req, "locationName", "location_name")
	handleFloat64OrNull(&setClauses, &args, req, "locationLat", "location_lat")
	handleFloat64OrNull(&setClauses, &args, req, "locationLong", "location_lng")

	if len(setClauses) > 0 {
		updateSQL := fmt.Sprintf("UPDATE posts SET %s WHERE post_id = ?", strings.Join(setClauses, ", "))
		args = append(args, postID)

		_, err = tx.Exec(updateSQL, args...)
		if err != nil {
			tx.Rollback()
			return models.UpdatePostResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to execute updateSQL - %v", err),
			}, err
		}
	}

	if imagesVal, ok := req["images"]; ok {
		if imagesVal == nil {
			_, err := tx.Exec("DELETE FROM post_media WHERE post_id = ?", postID)
			if err != nil {
				tx.Rollback()
				return models.UpdatePostResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to delete images - %v", err),
				}, err
			}
		} else {
			arr, ok := imagesVal.([]interface{})
			if !ok {
				tx.Rollback()
				return models.UpdatePostResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to update images. 'images' must be an array of strings."),
				}, fmt.Errorf("'images' must be an array of strings")
			}

			_, err := tx.Exec("DELETE FROM post_media WHERE post_id = ?", postID)
			if err != nil {
				tx.Rollback()
				return models.UpdatePostResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to delete old images due to the following error: %v", err),
				}, err
			}

			insertSQL := `INSERT INTO post_media (post_id, media_url, media_type) VALUES (?, ?, 'image')`
			stmt, err := tx.Prepare(insertSQL)
			if err != nil {
				tx.Rollback()
				return models.UpdatePostResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to prepare image insertion due to the following error: %v", err),
				}, err
			}
			defer stmt.Close()

			for _, imgVal := range arr {
				imgStr, ok := imgVal.(string)
				if !ok {
					tx.Rollback()
					return models.UpdatePostResponse{
						Success: false,
						Message: fmt.Sprintf("Failed to insert image. Image must be a string."),
					}, fmt.Errorf("image must be a string")
				}

				_, err := stmt.Exec(postID, imgStr, time.Now())
				if err != nil {
					tx.Rollback()
					return models.UpdatePostResponse{
						Success: false,
						Message: fmt.Sprintf("Failed to insert image due to the following error: %v", err),
					}, err
				}
			}
		}
	}

	if tagsVal, ok := req["hashtags"]; ok {
		if tagsVal == nil {
			_, err := tx.Exec("DELETE FROM post_hashtags WHERE post_id = ?", postID)
			if err != nil {
				tx.Rollback()
				return models.UpdatePostResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to delete 'hashtags' due to the following error: %v", err),
				}, err
			}
		} else {
			arr, ok := tagsVal.([]interface{})
			if !ok {
				tx.Rollback()
				return models.UpdatePostResponse{
					Success: false,
					Message: fmt.Sprintf("'hashtags' must be an array of strings"),
				}, fmt.Errorf("'hashtags' must be an array of strings")
			}

			_, err := tx.Exec("DELETE FROM post_hashtags WHERE post_id = ?", postID)
			if err != nil {
				tx.Rollback()
				return models.UpdatePostResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to delete 'hashtags' due to the following error: %v", err),
				}, err
			}

			insertTagSQL := `
              INSERT INTO hashtags (tag)
              VALUES (?) 
              ON DUPLICATE KEY UPDATE tag=VALUES(tag)
            `
			selectTagSQL := `SELECT hashtag_id FROM hashtags WHERE tag = ?`
			linkSQL := `INSERT INTO post_hashtags (post_id, hashtag_id) VALUES (?, ?)`

			for _, tVal := range arr {
				tStr, ok := tVal.(string)
				if !ok {
					tx.Rollback()
					return models.UpdatePostResponse{
						Success: false,
						Message: fmt.Sprintf("'tags' must be an array of strings"),
					}, fmt.Errorf("'tags' must be an array of strings")
				}
				cleanedTag := strings.TrimPrefix(tStr, "#")

				_, err := tx.Exec(insertTagSQL, cleanedTag)
				if err != nil {
					tx.Rollback()
					return models.UpdatePostResponse{
						Success: false,
						Message: fmt.Sprintf("Failed to insert tag due to the following error: %v", err),
					}, err
				}

				var tagID int64
				err = tx.QueryRow(selectTagSQL, cleanedTag).Scan(&tagID)
				if err != nil {
					tx.Rollback()
					return models.UpdatePostResponse{
						Success: false,
						Message: fmt.Sprintf("Failed to get tagID due to the following error: %v", err),
					}, err
				}
				_, err = tx.Exec(linkSQL, postID, tagID)
				if err != nil {
					tx.Rollback()
					return models.UpdatePostResponse{
						Success: false,
						Message: fmt.Sprintf("Failed to link tag due to the following error: %v", err),
					}, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return models.UpdatePostResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to commit the transaction due to the following error: %v", err),
		}, err
	}

	return models.UpdatePostResponse{
		Success: true,
		Message: "Successfully committed transaction and updated the post.",
		PostID:  postID,
	}, nil
}

func handleStringOrNull(setClauses *[]string, args *[]interface{}, req map[string]interface{}, jsonKey, columnName string) {
	val, ok := req[jsonKey]
	if !ok {
		return
	}
	if val == nil {
		*setClauses = append(*setClauses, fmt.Sprintf("%s = NULL", columnName))
		return
	}
	strVal, isString := val.(string)
	if !isString {
		log.Println("handleStringOrNull(val is not a string)")
		return
	}
	*setClauses = append(*setClauses, fmt.Sprintf("%s = ?", columnName))
	*args = append(*args, strVal)
}

func handleFloat64OrNull(setClauses *[]string, args *[]interface{}, req map[string]interface{}, jsonKey, columnName string) {
	val, ok := req[jsonKey]
	if !ok {
		return
	}
	if val == nil {
		*setClauses = append(*setClauses, fmt.Sprintf("%s = NULL", columnName))
		return
	}
	float64Val, isFloat := val.(float64)
	if !isFloat {
		log.Println("handleFloat64OrNull(val is not a float64)")
		return
	}
	*setClauses = append(*setClauses, fmt.Sprintf("%s = ?", columnName))
	*args = append(*args, float64Val)
}
