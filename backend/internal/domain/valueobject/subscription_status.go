package valueobject

import (
	"errors"
)

var (
	ErrInvalidSubscriptionStatus = errors.New("invalid subscription status")
)

type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusExpired   SubscriptionStatus = "expired"
	StatusCancelled SubscriptionStatus = "cancelled"
	StatusGrace     SubscriptionStatus = "grace"
)

// NewSubscriptionStatus creates a new SubscriptionStatus value object
func NewSubscriptionStatus(status string) (SubscriptionStatus, error) {
	s := SubscriptionStatus(status)
	switch s {
	case StatusActive, StatusExpired, StatusCancelled, StatusGrace:
		return s, nil
	default:
		return "", ErrInvalidSubscriptionStatus
	}
}

// String returns the string representation of the status
func (s SubscriptionStatus) String() string {
	return string(s)
}

// IsActive returns true if the status is active
func (s SubscriptionStatus) IsActive() bool {
	return s == StatusActive
}

// IsTerminated returns true if the subscription is cancelled or expired
func (s SubscriptionStatus) IsTerminated() bool {
	return s == StatusCancelled || s == StatusExpired
}
