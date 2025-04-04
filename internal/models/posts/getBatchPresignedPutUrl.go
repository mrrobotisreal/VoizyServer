package models

type GetBatchPresignedPutUrlRequest struct {
	UserID int64    `json:"userID"`
	PostID int64  `json:"postID"`
	FileNames []string `json:"fileNames"`
}

type PresignedFile struct {
	FileName 		 string `json:"fileName"`
	PresignedURL string `json:"presignedURL"`
	FinalURL 		 string `json:"finalURL"`
}

type GetBatchPresignedPutUrlResponse struct {
	Images []PresignedFile `json:"images"`
}
