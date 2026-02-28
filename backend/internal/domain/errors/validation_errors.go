package errors

import (
	"errors"
	"fmt"
)

var (
	// General validation errors
	ErrInvalidInput  = errors.New("invalid input")
	ErrRequiredField = errors.New("required field is missing")
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidLength = errors.New("invalid length")
	ErrOutOfRange    = errors.New("value out of range")

	// Specific field validation errors
	ErrInvalidEmail    = errors.New("invalid email address")
	ErrInvalidPlatform = errors.New("invalid platform")
	ErrInvalidPlanType = errors.New("invalid plan type")
	ErrInvalidStatus   = errors.New("invalid status")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrInvalidCurrency = errors.New("invalid currency code")
	ErrInvalidReceipt  = errors.New("invalid receipt data")
)

// ValidationError wraps a field validation error
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation failed for field '%s': %v", e.Field, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
