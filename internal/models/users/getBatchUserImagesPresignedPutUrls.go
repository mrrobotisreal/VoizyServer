package models

type GetBatchUserImagesPresignedPutUrlsRequest struct {
	UserID int64    `json:"userID"`
	FileNames []string `json:"fileNames"`
}

type PresignedFile struct {
	FileName 		 string `json:"fileName"`
	PresignedURL string `json:"presignedURL"`
	FinalURL 		 string `json:"finalURL"`
}

type GetBatchUserImagesPresignedPutUrlsResponse struct {
	Images []PresignedFile `json:"images"`
}
