package handlers

//
//import (
//	"VoizyServer/internal/database"
//	models "VoizyServer/internal/models/middleware"
//	"VoizyServer/internal/util"
//	"context"
//	"database/sql"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/golang-jwt/jwt/v5"
//	"net/http"
//	"strings"
//	"time"
//)
//
//func ValidateJWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		authHeader := r.Header.Get("Authorization")
//		if authHeader == "" {
//			sendError(w, "Authorization header missing", http.StatusUnauthorized)
//			return
//		}
//
//		splitToken := strings.Split(authHeader, "Bearer ")
//		if len(splitToken) != 2 {
//			sendError(w, "Invalid authorization format", http.StatusUnauthorized)
//			return
//		}
//
//		tokenStr := splitToken[1]
//
//		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
//			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
//			}
//			return []byte(util.GetJWTSecret()), nil
//		})
//
//		if err != nil {
//			sendError(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
//			return
//		}
//
//		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
//			if exp, ok := claims["exp"].(float64); ok {
//				if time.Now().Unix() > int64(exp) {
//					sendError(w, "Token has expired", http.StatusUnauthorized)
//					return
//				}
//			}
//
//			username, ok := claims["username"].(string)
//			if !ok {
//				sendError(w, "Invalid token claims", http.StatusUnauthorized)
//				return
//			}
//
//			ctx := context.WithValue(r.Context(), models.UsernameContextKey, username)
//			next.ServeHTTP(w, r.WithContext(ctx))
//		} else {
//			sendError(w, "Invalid token claims", http.StatusUnauthorized)
//			return
//		}
//	}
//}
//
//func ValidateAPIKeyMiddleware(next http.HandlerFunc) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		apiKey := r.Header.Get("X-API-Key")
//		if apiKey == "" {
//			sendError(w, "API key missing", http.StatusUnauthorized)
//			return
//		}
//
//		var user map[string]interface{}
//		query := `
//			SELECT
//				user_id,
//				api_key
//			FROM users
//			WHERE api_key = ?
//			LIMIT 1
//		`
//		row := database.DB.QueryRow(query, apiKey)
//		err := row.Scan(&user)
//		if err != nil {
//			if errors.Is(err, sql.ErrNoRows) {
//				return models.GetUserProfileResponse{}, fmt.Errorf("user profile not found: %w", err)
//			}
//			return models.GetUserProfileResponse{}, err
//		}
//
//		//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//		//defer cancel()
//		//
//		//collection := db.MongoClient.Database(db.DbName).Collection(db.UserCollection)
//		//var user types.User
//		//err := collection.FindOne(ctx, bson.M{"apiKey.key": apiKey}).Decode(&user)
//		//if err != nil {
//		//	sendError(w, "Invalid API key", http.StatusUnauthorized)
//		//	return
//		//}
//
//		if err := util.ValidateAPIKey(&user.APIKey); err != nil {
//			sendError(w, fmt.Sprintf("API key validation failed: %v", err), http.StatusUnauthorized)
//			return
//		}
//
//		limiter := util.NewRateLimiter().GetLimiter(apiKey)
//		if !limiter.Allow() {
//			sendError(w, "Rate limit exceeded", http.StatusTooManyRequests)
//			return
//		}
//
//		_, err = collection.UpdateOne(
//			ctx,
//			bson.M{"apiKey.key": apiKey},
//			bson.M{"$set": bson.M{"apiKey.last_used": time.Now()}},
//		)
//		if err != nil {
//			fmt.Printf("Error updating API key last used time: %v\n", err)
//		}
//
//		ctx = context.WithValue(r.Context(), models.APIKeyContextKey, apiKey)
//		ctx = context.WithValue(ctx, models.UsernameContextKey, user.Username)
//		next.ServeHTTP(w, r.WithContext(ctx))
//	}
//}
//
//func CombinedAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
//	return ValidateJWTMiddleware(ValidateAPIKeyMiddleware(next))
//}
//
//func sendError(w http.ResponseWriter, message string, status int) {
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(status)
//	json.NewEncoder(w).Encode(models.ErrorResponse{
//		Error:   http.StatusText(status),
//		Message: message,
//	})
//}
//
//func GetUsernameFromContext(ctx context.Context) (string, bool) {
//	username, ok := ctx.Value(models.UsernameContextKey).(string)
//	return username, ok
//}
//
//func GetAPIKeyFromContext(ctx context.Context) (string, bool) {
//	apiKey, ok := ctx.Value(models.APIKeyContextKey).(string)
//	return apiKey, ok
//}
