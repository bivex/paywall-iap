package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// GracePeriodStatus represents the status of a grace period
type GracePeriodStatus string

const (
	GraceStatusActive   GracePeriodStatus = "active"
	GraceStatusResolved GracePeriodStatus = "resolved"
	GraceStatusExpired  GracePeriodStatus = "expired"
)

// GracePeriod represents a grace period for subscription renewal failures
type GracePeriod struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	SubscriptionID uuid.UUID
	Status         GracePeriodStatus
	ExpiresAt      time.Time
	ResolvedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewGracePeriod creates a new active grace period
func NewGracePeriod(userID, subscriptionID uuid.UUID, expiresAt time.Time) *GracePeriod {
	now := time.Now()
	return &GracePeriod{
		ID:             uuid.New(),
		UserID:         userID,
		SubscriptionID: subscriptionID,
		Status:         GraceStatusActive,
		ExpiresAt:      expiresAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// IsActive returns true if the grace period is currently active
func (gp *GracePeriod) IsActive() bool {
	return gp.Status == GraceStatusActive && gp.ExpiresAt.After(time.Now())
}

// IsExpired returns true if the grace period has expired
func (gp *GracePeriod) IsExpired() bool {
	return gp.Status == GraceStatusExpired || gp.ExpiresAt.Before(time.Now())
}

// IsResolved returns true if the grace period has been resolved
func (gp *GracePeriod) IsResolved() bool {
	return gp.Status == GraceStatusResolved
}

// CanAccessContent returns true if user can access content during grace period
func (gp *GracePeriod) CanAccessContent() bool {
	return gp.IsActive()
}

// Resolve marks the grace period as resolved (user renewed successfully)
func (gp *GracePeriod) Resolve() error {
	if gp.Status == GraceStatusExpired {
		return errors.New("cannot resolve expired grace period")
	}

	now := time.Now()
	gp.Status = GraceStatusResolved
	gp.ResolvedAt = &now
	gp.UpdatedAt = now
	return nil
}

// Expire marks the grace period as expired
func (gp *GracePeriod) Expire() error {
	if gp.Status == GraceStatusResolved {
		return errors.New("cannot expire already resolved grace period")
	}

	gp.Status = GraceStatusExpired
	gp.UpdatedAt = time.Now()
	return nil
}

// DaysRemaining returns the number of days remaining in the grace period
func (gp *GracePeriod) DaysRemaining() int {
	if gp.IsExpired() {
		return 0
	}

	duration := gp.ExpiresAt.Sub(time.Now())
	if duration < 0 {
		return 0
	}

	return int(duration.Hours() / 24)
}

// HoursRemaining returns the number of hours remaining in the grace period
func (gp *GracePeriod) HoursRemaining() int {
	if gp.IsExpired() {
		return 0
	}

	duration := gp.ExpiresAt.Sub(time.Now())
	if duration < 0 {
		return 0
	}

	return int(duration.Hours())
}
