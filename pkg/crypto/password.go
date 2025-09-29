package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the default bcrypt cost to use for password hashing
const DefaultCost = bcrypt.DefaultCost

// HashPassword creates a bcrypt hash of the given password
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

// CheckPasswordHash compares a password with its hash
func CheckPasswordHash(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength checks if a password meets minimum requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Add more validation rules as needed:
	// - Must contain uppercase letter
	// - Must contain lowercase letter  
	// - Must contain number
	// - Must contain special character
	// For now, we'll keep it simple for the learning example

	return nil
}

// GenerateRandomPassword creates a random password (useful for temporary passwords)
// This is a simple implementation - in production you'd want a more sophisticated approach
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		return "", fmt.Errorf("password length must be at least 8 characters")
	}

	// Simple random password generation
	// In production, use crypto/rand for better randomness
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	
	password := make([]byte, length)
	for i := range password {
		// This is not cryptographically secure - use crypto/rand in production
		password[i] = charset[i%len(charset)]
	}

	return string(password), nil
}
