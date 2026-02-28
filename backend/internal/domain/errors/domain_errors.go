package errors

import (
	"errors"
	"fmt"
)

var (
	// User errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	// Subscription errors
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrSubscriptionNotActive     = errors.New("subscription is not active")
	ErrSubscriptionExpired       = errors.New("subscription has expired")
	ErrSubscriptionCancelled     = errors.New("subscription has been cancelled")
	ErrActiveSubscriptionExists  = errors.New("active subscription already exists")

	// Transaction errors
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrDuplicateReceipt     = errors.New("receipt has already been processed")
	ErrReceiptInvalid       = errors.New("receipt is invalid")
	ErrReceiptExpired       = errors.New("receipt has expired")

	// Payment errors
	ErrPaymentFailed   = errors.New("payment failed")
	ErrPaymentRefunded = errors.New("payment has been refunded")

	// External service errors
	ErrExternalServiceUnavailable = errors.New("external service unavailable")
	ErrIAPVerificationFailed     = errors.New("IAP verification failed")
)

// NotFoundError wraps an error with not found context
type NotFoundError struct {
	Entity string
	ID     string
	Err    error
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id '%s' not found: %v", e.Entity, e.ID, e.Err)
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

// ConflictError wraps an error with conflict context
type ConflictError struct {
	Entity string
	Reason string
	Err    error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s conflict: %s - %v", e.Entity, e.Reason, e.Err)
}

func (e *ConflictError) Unwrap() error {
	return e.Err
}
