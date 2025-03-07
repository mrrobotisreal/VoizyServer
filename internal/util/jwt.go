package util

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const jwtSecretKey = "voizy"

func GetJWTSecret() string {
	return jwtSecretKey
}

func GenerateAndStoreJWT(username, sessionOption string) (string, error) {
	var expirationTime time.Time

	switch sessionOption {
	case "always":
		expirationTime = time.Now().Add(366 * 244 * time.Hour)
	case "daily":
		expirationTime = time.Now().Add(24 * time.Hour)
	case "weekly":
		expirationTime = time.Now().Add(7 * 24 * time.Hour)
	case "monthly":
		expirationTime = time.Now().Add(30 * 24 * time.Hour)
	case "never":
		expirationTime = time.Now().Add(1 * time.Minute)
	default:
		return "", errors.New("invalid session option")
	}

	claims := jwt.MapClaims{
		"username": username,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}

	return tokenString, nil
}
