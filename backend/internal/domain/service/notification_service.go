package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
)

// NotificationService handles sending notifications to users
type NotificationService struct {
	// In production, this would integrate with email/SMS/push notification services
	// For now, we'll log notifications
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// SendGracePeriodExpiringNotification sends a notification when grace period is expiring soon
func (s *NotificationService) SendGracePeriodExpiringNotification(ctx context.Context, userID uuid.UUID, gracePeriod *entity.GracePeriod) error {
	// In production, send email/push notification
	// For now, log the notification
	fmt.Printf("[NOTIFICATION] User %s: Grace period expiring in %d hours for subscription %s\n",
		userID, gracePeriod.HoursRemaining(), gracePeriod.SubscriptionID)

	// TODO: Integrate with email service (SendGrid, SES)
	// TODO: Integrate with push notification service (Firebase, APNs)

	return nil
}

// SendWinbackOfferNotification sends a winback offer to churned users
func (s *NotificationService) SendWinbackOfferNotification(ctx context.Context, userID uuid.UUID, offer *entity.WinbackOffer) error {
	fmt.Printf("[NOTIFICATION] User %s: Winback offer %s - %.2f discount available\n",
		userID, offer.CampaignID, offer.DiscountValue)

	return nil
}

// SendSubscriptionExpiredNotification sends notification when subscription expires
func (s *NotificationService) SendSubscriptionExpiredNotification(ctx context.Context, userID uuid.UUID, subscriptionID uuid.UUID) error {
	fmt.Printf("[NOTIFICATION] User %s: Subscription %s has expired\n", userID, subscriptionID)
	return nil
}
