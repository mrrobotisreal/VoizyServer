package models

type Song struct {
	SongID  int64  `json:"songID"`
	Title 	string `json:"title"`
	Artist 	string `json:"artist"`
	SongURL string `json:"songURL"`
}

type ListSongsResponse struct {
	Songs 		 []Song `json:"songs"`
	Limit 		 int64 	`json:"limit"`
	Page 			 int64 	`json:"page"`
	TotalSongs int64 	`json:"totalSongs"`
	TotalPages int64 	`json:"totalPages"`
}
