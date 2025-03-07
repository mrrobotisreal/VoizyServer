package main

import (
	"VoizyServer/internal/database"
	analyticsHandlers "VoizyServer/internal/handlers/analytics"
	authHandlers "VoizyServer/internal/handlers/auth"
	postHandlers "VoizyServer/internal/handlers/posts"
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
	http.HandleFunc("/users/profile/list", userHandlers.ListUserProfilesHandler) // temp handler
	http.HandleFunc("/users/profile/update", userHandlers.UpdateUserProfileHandler)

	// POSTS
	http.HandleFunc("/posts/create", postHandlers.CreatePostHandler)
	http.HandleFunc("/posts/list", postHandlers.ListPostsHandler)
	http.HandleFunc("/posts/get/total", postHandlers.GetTotalPostsHandler)
	http.HandleFunc("/posts/get/details", postHandlers.GetPostDetailsHandler)
	http.HandleFunc("/posts/get/media", postHandlers.GetPostMediaHandler)
	http.HandleFunc("/posts/reactions/put", postHandlers.PutPostReactionHandler)
	http.HandleFunc("/posts/comments/put", postHandlers.PutCommentHandler)
	http.HandleFunc("/posts/comments/list", postHandlers.ListPostCommentsHandler)
	http.HandleFunc("/posts/comments/reactions/put", postHandlers.PutCommentReactionHandler)

	// ANALYTICS
	http.HandleFunc("/analytics/track", analyticsHandlers.BatchTrackEventsHandler)
	http.HandleFunc("/analytics/events/list", analyticsHandlers.ListEventsHandler)
	http.HandleFunc("/analytics/stats/list", analyticsHandlers.ListStatsHandler)

	fmt.Println("Server running on localhost:9295")
	if err := http.ListenAndServe(":9295", nil); err != nil {
		log.Fatal(err)
	}
}
