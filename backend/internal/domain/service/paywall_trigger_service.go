package service

import (
	"context"

	"github.com/bivex/paywall-iap/internal/domain/repository"
	"github.com/google/uuid"
)

// TriggerStatus represents the paywall trigger evaluation result
type TriggerStatus struct {
	ShouldShowPaywall     bool    `json:"should_show_paywall"`
	ShowD2CButton         bool    `json:"show_d2c_button"`
	TriggerReason         string  `json:"trigger_reason"`
	SessionCount          int     `json:"session_count"`
	HasActiveSubscription bool    `json:"has_active_subscription"`
	PurchaseChannel       *string `json:"purchase_channel"`
}

// PaywallTriggerService evaluates when to show the paywall
type PaywallTriggerService struct {
	userRepo         repository.UserRepository
	subscriptionRepo repository.SubscriptionRepository
}

func NewPaywallTriggerService(userRepo repository.UserRepository, subscriptionRepo repository.SubscriptionRepository) *PaywallTriggerService {
	return &PaywallTriggerService{userRepo: userRepo, subscriptionRepo: subscriptionRepo}
}

// Evaluate returns the trigger status for a given user
func (s *PaywallTriggerService) Evaluate(ctx context.Context, userID uuid.UUID) (*TriggerStatus, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if user has active subscription
	hasActiveSub := false
	_, subErr := s.subscriptionRepo.GetActiveByUserID(ctx, userID)
	if subErr == nil {
		hasActiveSub = true
	}

	status := &TriggerStatus{
		SessionCount:          user.SessionCount,
		HasActiveSubscription: hasActiveSub,
		PurchaseChannel:       user.PurchaseChannel,
		ShowD2CButton:         user.ShouldShowD2CButton(),
	}

	// Don't show paywall if user already has active subscription
	if hasActiveSub {
		status.ShouldShowPaywall = false
		status.TriggerReason = "has_active_subscription"
		return status, nil
	}

	// Trigger logic based on session count and behavior
	if user.SessionCount >= 3 {
		status.ShouldShowPaywall = true
		status.TriggerReason = "session_threshold"
	} else if user.HasViewedAds && user.SessionCount >= 2 {
		status.ShouldShowPaywall = true
		status.TriggerReason = "viewed_ads"
	} else {
		status.ShouldShowPaywall = false
		status.TriggerReason = "new_user"
	}

	return status, nil
}

