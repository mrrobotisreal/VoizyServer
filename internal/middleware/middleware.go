package middleware

import (
	"VoizyServer/internal/database"
	models "VoizyServer/internal/models/middleware"
	"VoizyServer/internal/util"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ValidateJWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendError(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) != 2 {
			sendError(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenStr := splitToken[1]

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(util.GetJWTSecret()), nil
		})

		if err != nil {
			sendError(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					sendError(w, "Token has expired", http.StatusUnauthorized)
					return
				}
			}

			userID, ok := claims["userID"].(string)
			if !ok {
				log.Println("JWT userID: ", userID)
				sendError(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), models.UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			sendError(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
	}
}

func ValidateAPIKeyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		xApiKey := r.Header.Get("X-API-Key")
		if xApiKey == "" {
			sendError(w, "X-API-Key key missing", http.StatusUnauthorized)
			return
		}
		xUserIDString := r.Header.Get("X-User-ID")
		if xUserIDString == "" {
			sendError(w, "X-User-ID missing", http.StatusUnauthorized)
			return
		}
		xUserID, err := strconv.ParseInt(xUserIDString, 10, 64)
		if err != nil {
			log.Println("Failed to parse xUserIDString (string) to xUserID (int64) due to the following error: ", err)
			sendError(w, "Failed to parse X-User-ID.", http.StatusInternalServerError)
			return
		}

		var apiKeyUserID int64
		var apiKey models.APIKey
		apiKeyQuery := `
			SELECT
				user_id,
				api_key,
				created_at,
				expires_at,
				last_used_at,
				updated_at
			FROM api_keys
			WHERE user_id = ? AND api_key = ?
			LIMIT 1
		`
		row := database.DB.QueryRow(apiKeyQuery, xUserID, xApiKey)
		err = row.Scan(&apiKeyUserID,
			&apiKey.Key,
			&apiKey.CreatedAt,
			&apiKey.ExpiresAt,
			&apiKey.LastUsedAt,
			&apiKey.UpdatedAt,
		)
		if err != nil {
			log.Println("Scan row error: ", err)
			sendError(w, "Failed to find apiKey or user", http.StatusUnauthorized)
			return
		}

		var userID int64
		var userApiKey string
		userQuery := `SELECT user_id, api_key FROM users WHERE user_id = ? AND api_key = ?`
		row = database.DB.QueryRow(userQuery, xUserID, xApiKey)
		err = row.Scan(&userID, &userApiKey)
		if err != nil {
			log.Println("Scan row error: ", err)
			sendError(w, "Failed to find user or apiKey", http.StatusUnauthorized)
			return
		}

		if err := util.ValidateAPIKey(&apiKey); err != nil {
			sendError(w, fmt.Sprintf("API key validation failed: %v", err), http.StatusUnauthorized)
			return
		}

		limiter := util.NewRateLimiter().GetLimiter(apiKey.Key)
		if !limiter.Allow() {
			sendError(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		err = updateAPIKeyLastUsedAt(userID, apiKey.Key)
		if err != nil {
			sendError(w, "Failed to update API Key usage.", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), models.APIKeyContextKey, apiKey)
		ctx = context.WithValue(ctx, models.UserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func CombinedAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return ValidateJWTMiddleware(ValidateAPIKeyMiddleware(next))
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(models.UserIDContextKey).(string)
	return username, ok
}

func GetAPIKeyFromContext(ctx context.Context) (string, bool) {
	apiKey, ok := ctx.Value(models.APIKeyContextKey).(string)
	return apiKey, ok
}

func updateAPIKeyLastUsedAt(userID int64, apiKey string) error {
	query := `
		UPDATE api_keys
		SET last_used_at = NOW()
		WHERE user_id = ? AND api_key = ?
	`
	updateResult, err := database.DB.Exec(query, userID, apiKey)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to update last_used_at for api_key = %s and user_id = %d due to the following error: %v", apiKey, userID, err))
		return err
	}
	rowsAffected, _ := updateResult.RowsAffected()
	if rowsAffected != 1 {
		log.Println("Incorrect rows affected: ", rowsAffected)
		return fmt.Errorf("rows affected should have been 1 but was %d", rowsAffected)
	}

	return nil
}
