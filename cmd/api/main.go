package main

import (
	"VoizyServer/internal/database"
	authHandlers "VoizyServer/internal/handlers/auth"
	userHandlers "VoizyServer/internal/handlers/users"
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
	http.HandleFunc("/users/create", userHandlers.CreateUserHandler)
	http.HandleFunc("/users/login", authHandlers.LoginHandler)
	http.HandleFunc("/users/get", userHandlers.GetUserHandler)
	http.HandleFunc("/users/update", userHandlers.UpdateUserHandler)
	http.HandleFunc("/users/profile/get", userHandlers.GetProfileHandler)
	http.HandleFunc("/users/profile/update", userHandlers.UpdateUserProfileHandler)

	// POSTS
	//http.HandleFunc("/posts/create", nil)

	fmt.Println("Server running on localhost:9295")
	if err := http.ListenAndServe(":9295", nil); err != nil {
		log.Fatal(err)
	}
}
