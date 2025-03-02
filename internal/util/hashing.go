package util

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
)

// Generate a new salt
func GenerateSalt(length int) (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode to Base64
	return base64.StdEncoding.EncodeToString(randomBytes), nil
}

// Hash password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 15)

	return string(bytes), err
}

// Check password hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}
