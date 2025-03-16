package handlers

import (
	aws "VoizyServer/internal/aws"
	models "VoizyServer/internal/models/users"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func GetBatchUserImagesPresignedPutUrlsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	var req models.GetBatchUserImagesPresignedPutUrlsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		return
	}

	response, err := getBatchUserImagesPresignedPutUrls(req)
	if err != nil {
		log.Println("Failed to get presigned URLs due to the following error: ", err)
		http.Error(w, "Failed to get presigned URLs.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getBatchUserImagesPresignedPutUrls(req models.GetBatchUserImagesPresignedPutUrlsRequest) (models.GetBatchUserImagesPresignedPutUrlsResponse, error) {
	var response models.GetBatchUserImagesPresignedPutUrlsResponse
	presignClient := s3.NewPresignClient(aws.S3Client)
	bucket := "voizy-app"

	var results []models.PresignedFile
	for _, fileName := range req.FileNames {
		if fileName == "" {
			continue
		}
		key := fmt.Sprintf("%d/%s/%s", req.UserID, "photos", fileName)
		input := &s3.PutObjectInput{
			Bucket: &bucket,
			Key: &key,
		}
		presignReq, err := presignClient.PresignPutObject(
			context.TODO(),
			input,
			s3.WithPresignExpires(5*time.Minute),
		)
		if err != nil {
			log.Println("Failed to presign put object:", key, "err:", err)
			continue
		}

		finalURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)

		results = append(results, models.PresignedFile{
			FileName:    key,
			PresignedURL: presignReq.URL,
			FinalURL:    finalURL,
		})
	}
	response.Images = results

	return response, nil
}
