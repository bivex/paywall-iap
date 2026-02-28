package entity

import (
	"time"

	"github.com/google/uuid"
)

// DunningStatus represents the status of a dunning process
type DunningStatus string

const (
	DunningStatusPending    DunningStatus = "pending"
	DunningStatusInProgress DunningStatus = "in_progress"
	DunningStatusRecovered  DunningStatus = "recovered"
	DunningStatusFailed     DunningStatus = "failed"
)

// Dunning represents a dunning process for failed subscription renewal
type Dunning struct {
	ID             uuid.UUID
	SubscriptionID uuid.UUID
	UserID         uuid.UUID
	Status         DunningStatus
	AttemptCount   int
	MaxAttempts    int
	NextAttemptAt  time.Time
	LastAttemptAt  *time.Time
	RecoveredAt    *time.Time
	FailedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewDunning creates a new dunning process
func NewDunning(subscriptionID, userID uuid.UUID, nextAttemptAt time.Time) *Dunning {
	now := time.Now()
	return &Dunning{
		ID:             uuid.New(),
		SubscriptionID: subscriptionID,
		UserID:         userID,
		Status:         DunningStatusPending,
		AttemptCount:   0,
		MaxAttempts:    5,
		NextAttemptAt:  nextAttemptAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// CanRetry returns true if dunning can be retried
func (d *Dunning) CanRetry() bool {
	return d.Status == DunningStatusPending || d.Status == DunningStatusInProgress
}

// IsRecovered returns true if dunning was successful
func (d *Dunning) IsRecovered() bool {
	return d.Status == DunningStatusRecovered
}

// IsFailed returns true if dunning failed
func (d *Dunning) IsFailed() bool {
	return d.Status == DunningStatusFailed
}

// IncrementAttempt increments the attempt counter
func (d *Dunning) IncrementAttempt() {
	d.AttemptCount++
	now := time.Now()
	d.LastAttemptAt = &now
	d.UpdatedAt = now
}

// MarkRecovered marks the dunning as recovered
func (d *Dunning) MarkRecovered() {
	d.Status = DunningStatusRecovered
	now := time.Now()
	d.RecoveredAt = &now
	d.UpdatedAt = now
}

// MarkFailed marks the dunning as failed
func (d *Dunning) MarkFailed() {
	d.Status = DunningStatusFailed
	now := time.Now()
	d.FailedAt = &now
	d.UpdatedAt = now
}

// GetRetryDelay returns the delay before next retry based on attempt count
func (d *Dunning) GetRetryDelay() time.Duration {
	// Exponential backoff: 1 day, 3 days, 7 days, 14 days, 30 days
	switch d.AttemptCount {
	case 0:
		return 24 * time.Hour
	case 1:
		return 72 * time.Hour
	case 2:
		return 168 * time.Hour
	case 3:
		return 336 * time.Hour
	default:
		return 720 * time.Hour
	}
}
