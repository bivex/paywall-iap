package valueobject

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
)

// Email represents a valid email address
type Email struct {
	value string
}

// NewEmail creates a new Email value object
func NewEmail(email string) (*Email, error) {
	if !isValidEmail(email) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidEmail, email)
	}
	return &Email{value: email}, nil
}

// String returns the email string
func (e *Email) String() string {
	return e.value
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	// Simple email regex for validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
