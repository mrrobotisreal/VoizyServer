package main

import (
	"VoizyServer/internal/database"
	"fmt"
	"log"
	"net/http"
)

func main() {
	if err := database.InitMySQL(); err != nil {
		log.Fatalf("Failed to init MySQL: %v", err)
	}
	defer database.DB.Close()

	if err := database.InitRedis(); err != nil {
		log.Fatalf("Failed to init Redis: %v", err)
	}
	defer database.RDB.Close()

	// USERS
	http.HandleFunc("/users/create", nil)

	// POSTS
	http.HandleFunc("/posts/create", nil)

	fmt.Println("Server running on localhost:9295")
	if err := http.ListenAndServe(":9295", nil); err != nil {
		log.Fatal(err)
	}
}
