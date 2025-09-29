package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field, message string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	
	// Simple email regex - in production you might want a more sophisticated one
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	
	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	
	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	
	// Check for at least one number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	
	return nil
}

// ValidateName validates user name
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	
	if len(strings.TrimSpace(name)) < 2 {
		return fmt.Errorf("name must be at least 2 characters long")
	}
	
	if len(name) > 100 {
		return fmt.Errorf("name must be less than 100 characters")
	}
	
	return nil
}

// ValidateUserRegistration validates a complete user registration request
func ValidateUserRegistration(name, email, password string) ValidationErrors {
	var errors ValidationErrors
	
	if err := ValidateName(name); err != nil {
		errors.Add("name", err.Error())
	}
	
	if err := ValidateEmail(email); err != nil {
		errors.Add("email", err.Error())
	}
	
	if err := ValidatePassword(password); err != nil {
		errors.Add("password", err.Error())
	}
	
	return errors
}
