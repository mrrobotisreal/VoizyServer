package main

import (
	"VoizyServer/internal/database"
	analyticsHandlers "VoizyServer/internal/handlers/analytics"
	authHandlers "VoizyServer/internal/handlers/auth"
	postHandlers "VoizyServer/internal/handlers/posts"
	userHandlers "VoizyServer/internal/handlers/users"
	"VoizyServer/internal/middleware"
	"fmt"
	"log"
	"net/http"
)

func main() {
	if err := database.InitMySQL(); err != nil {
		log.Fatalf("Failed to init MySQL: %v", err)
	}
	defer database.DB.Close()

	//if err := database.InitRedis(); err != nil {
	//	log.Fatalf("Failed to init Redis: %v", err)
	//}
	//defer database.RDB.Close()

	/// USERS ///
	// Create and Login
	http.HandleFunc("/users/create", userHandlers.CreateUserHandler)
	http.HandleFunc("/users/login", authHandlers.LoginHandler)
	// User
	http.HandleFunc("/users/get", middleware.ValidateAPIKeyMiddleware(userHandlers.GetUserHandler))
	http.HandleFunc("/users/update", middleware.CombinedAuthMiddleware(userHandlers.UpdateUserHandler))
	// User Profile
	http.HandleFunc("/users/profile/get", middleware.ValidateAPIKeyMiddleware(userHandlers.GetProfileHandler))
	http.HandleFunc("/users/profile/list", middleware.ValidateAPIKeyMiddleware(userHandlers.ListUserProfilesHandler))
	http.HandleFunc("/users/profile/update", middleware.CombinedAuthMiddleware(userHandlers.UpdateUserProfileHandler))
	// User Songs
	http.HandleFunc("/songs/list", middleware.ValidateAPIKeyMiddleware(userHandlers.ListSongsHandler))
	// User Images
	http.HandleFunc("/users/images/get/total", middleware.ValidateAPIKeyMiddleware(userHandlers.GetTotalImages))
	http.HandleFunc("/users/images/get/profilePic", middleware.ValidateAPIKeyMiddleware(userHandlers.GetProfilePicHandler))
	http.HandleFunc("/users/images/get/coverPic", middleware.ValidateAPIKeyMiddleware(userHandlers.GetCoverPicHandler))
	http.HandleFunc("/users/images/list", middleware.ValidateAPIKeyMiddleware(userHandlers.ListImagesHandler))
	http.HandleFunc("/users/images/profilePic/update", middleware.CombinedAuthMiddleware(userHandlers.UpdateProfilePicHandler))
	http.HandleFunc("/users/images/coverPic/update", middleware.CombinedAuthMiddleware(userHandlers.UpdateCoverPicHandler))
	http.HandleFunc("/users/images/put", middleware.CombinedAuthMiddleware(userHandlers.PutUserImagesHandler))
	http.HandleFunc("/users/images/batch/get/presigned", middleware.CombinedAuthMiddleware(userHandlers.GetBatchUserImagesPresignedPutUrlsHandler))
	// Friendships
	http.HandleFunc("/users/friends/create", middleware.CombinedAuthMiddleware(userHandlers.CreateFriendRequestHandler))
	http.HandleFunc("/users/friends/list", middleware.ValidateAPIKeyMiddleware(userHandlers.ListFriendshipsHandler))
	http.HandleFunc("/users/friends/list/common", middleware.ValidateAPIKeyMiddleware(userHandlers.ListFriendsInCommonHandler))
	http.HandleFunc("/users/friends/get/total", middleware.ValidateAPIKeyMiddleware(userHandlers.GetTotalFriendsHandler))

	/// POSTS ///
	// Posts
	http.HandleFunc("/posts/create", middleware.CombinedAuthMiddleware(postHandlers.CreatePostHandler))
	// http.HandleFunc("/posts/get/presigned", middleware.CombinedAuthMiddleware(postHandlers.GetPresignedPutUrlHandler))
	http.HandleFunc("/posts/batch/get/presigned", middleware.CombinedAuthMiddleware(postHandlers.GetBatchPresignedPutUrlHandler))
	http.HandleFunc("/posts/update", middleware.CombinedAuthMiddleware(postHandlers.UpdatePostHandler))
	http.HandleFunc("/posts/list", middleware.ValidateAPIKeyMiddleware(postHandlers.ListPostsHandler))
	http.HandleFunc("/posts/get/total", middleware.ValidateAPIKeyMiddleware(postHandlers.GetTotalPostsHandler))
	http.HandleFunc("/posts/get/details", middleware.ValidateAPIKeyMiddleware(postHandlers.GetPostDetailsHandler))
	http.HandleFunc("/posts/get/media", middleware.ValidateAPIKeyMiddleware(postHandlers.GetPostMediaHandler))
	http.HandleFunc("/posts/put/media", middleware.CombinedAuthMiddleware(postHandlers.PutPostMediaHandler))
	http.HandleFunc("/posts/reactions/put", middleware.CombinedAuthMiddleware(postHandlers.PutPostReactionHandler))
	// Comments
	http.HandleFunc("/posts/comments/get/total", middleware.ValidateAPIKeyMiddleware(postHandlers.GetTotalCommentsHandler))
	http.HandleFunc("/posts/comments/put", middleware.CombinedAuthMiddleware(postHandlers.PutCommentHandler))
	http.HandleFunc("/posts/comments/list", middleware.ValidateAPIKeyMiddleware(postHandlers.ListPostCommentsHandler))
	http.HandleFunc("/posts/comments/reactions/put", middleware.CombinedAuthMiddleware(postHandlers.PutCommentReactionHandler))
	// Feeds
	http.HandleFunc("/posts/feed/list", middleware.ValidateAPIKeyMiddleware(postHandlers.ListFeedHandler))
	http.HandleFunc("/posts/feed/recommended/get", middleware.ValidateAPIKeyMiddleware(postHandlers.GetRecommendedFeed))
	http.HandleFunc("/posts/feed/popular/get", middleware.ValidateAPIKeyMiddleware(postHandlers.GetPopularPosts))
	http.HandleFunc("/posts/feed/friends/get", middleware.ValidateAPIKeyMiddleware(postHandlers.GetFriendFeed))
	//http.HandleFunc("/posts/feed/recommended/list", middleware.ValidateAPIKeyMiddleware(postHandlers.ListRecommendedFeedHandler))
	// Impressions and Views
	http.HandleFunc("/posts/impressions/put", middleware.ValidateAPIKeyMiddleware(postHandlers.PutPostImpressionHandler))
	http.HandleFunc("/posts/views/put", middleware.ValidateAPIKeyMiddleware(postHandlers.PutPostViewHandler))

	/// ANALYTICS
	http.HandleFunc("/analytics/track", middleware.CombinedAuthMiddleware(analyticsHandlers.BatchTrackEventsHandler))
	http.HandleFunc("/analytics/events/list", middleware.CombinedAuthMiddleware(analyticsHandlers.ListEventsHandler))
	http.HandleFunc("/analytics/stats/list", middleware.CombinedAuthMiddleware(analyticsHandlers.ListStatsHandler))

	/// AUTH
	// http.HandleFunc("/api/keys/insert", authHandlers.InsertApiKeyHandler)

	certFile := "/etc/letsencrypt/live/voizy.me/fullchain.pem"
	keyFile := "/etc/letsencrypt/live/voizy.me/privkey.pem"

	fmt.Println("Server running securely on localhost:443")
	if err := http.ListenAndServeTLS(":443", certFile, keyFile, nil); err != nil {
		log.Fatal(err)
	}
	// fmt.Println("Server running securely on localhost:9295")
	// if err := http.ListenAndServe(":9295", nil); err != nil {
	// 	log.Fatal(err)
	// }
}
