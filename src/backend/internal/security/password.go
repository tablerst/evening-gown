package security

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrWeakPassword = errors.New("password too weak")

// HashPassword hashes a plaintext password using bcrypt.
func HashPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 10 {
		return "", ErrWeakPassword
	}
	// bcrypt handles salting internally.
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// CheckPassword verifies a plaintext password against a bcrypt hash.
func CheckPassword(hash, password string) bool {
	if strings.TrimSpace(hash) == "" {
		return false
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
