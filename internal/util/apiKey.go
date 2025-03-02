package util

import (
	models "VoizyServer/internal/models/middleware"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/time/rate"
	"sync"
	"time"
)

const (
	apiKeyLength    = 32
	maxRequestRate  = 100
	keyRotationDays = 90
)

type RateLimiter struct {
	limiters sync.Map
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

func (rl *RateLimiter) GetLimiter(apiKey string) *rate.Limiter {
	limiter, exists := rl.limiters.Load(apiKey)
	if !exists {
		newLimiter := rate.NewLimiter(rate.Limit(maxRequestRate), maxRequestRate)
		rl.limiters.Store(apiKey, newLimiter)
		return newLimiter
	}
	return limiter.(*rate.Limiter)
}

func GenerateSecureAPIKey() (*models.APIKey, error) {
	randomBytes := make([]byte, apiKeyLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("error generating random bytes: %v", err)
	}

	key := fmt.Sprintf("sk_%s", hex.EncodeToString(randomBytes))

	apiKey := &models.APIKey{
		Key:       key,
		Created:   time.Now(),
		LastUsed:  time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(keyRotationDays) * 24 * time.Hour),
	}

	return apiKey, nil
}

func ValidateAPIKey(apiKey *models.APIKey) error {
	if apiKey == nil {
		return errors.New("api key is nil")
	}

	if time.Now().After(apiKey.ExpiresAt) {
		return errors.New("api key has expired")
	}

	apiKey.LastUsed = time.Now()

	return nil
}

func IsKeyRotationNeeded(apiKey *models.APIKey) bool {
	return time.Now().After(apiKey.Created.Add(time.Duration(keyRotationDays) * 24 * time.Hour))
}
