package entity

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusExpired   SubscriptionStatus = "expired"
	StatusCancelled SubscriptionStatus = "cancelled"
	StatusGrace     SubscriptionStatus = "grace"
)

type SubscriptionSource string

const (
	SourceIAP    SubscriptionSource = "iap"
	SourceStripe SubscriptionSource = "stripe"
	SourcePaddle SubscriptionSource = "paddle"
)

type PlanType string

const (
	PlanMonthly  PlanType = "monthly"
	PlanAnnual   PlanType = "annual"
	PlanLifetime PlanType = "lifetime"
)

type Subscription struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Status    SubscriptionStatus
	Source    SubscriptionSource
	Platform  string
	ProductID string
	PlanType  PlanType
	ExpiresAt time.Time
	AutoRenew bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewSubscription creates a new subscription entity
func NewSubscription(userID uuid.UUID, source SubscriptionSource, platform, productID string, planType PlanType, expiresAt time.Time) *Subscription {
	return &Subscription{
		ID:        uuid.New(),
		UserID:    userID,
		Status:    StatusActive,
		Source:    source,
		Platform:  platform,
		ProductID: productID,
		PlanType:  planType,
		ExpiresAt: expiresAt,
		AutoRenew: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// IsActive returns true if the subscription is currently active
func (s *Subscription) IsActive() bool {
	if s.DeletedAt != nil {
		return false
	}
	return s.Status == StatusActive && s.ExpiresAt.After(time.Now())
}

// IsExpired returns true if the subscription has expired
func (s *Subscription) IsExpired() bool {
	return s.Status == StatusExpired || s.ExpiresAt.Before(time.Now())
}

// CanAccessContent returns true if user can access premium content
func (s *Subscription) CanAccessContent() bool {
	if s.DeletedAt != nil {
		return false
	}
	return s.IsActive()
}

// HasGracePeriod returns true if subscription is in grace period
func (s *Subscription) HasGracePeriod() bool {
	return s.Status == StatusGrace
}
