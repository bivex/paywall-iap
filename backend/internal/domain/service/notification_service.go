package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
)

// NotificationService handles sending notifications to users via SendGrid (email) and FCM (push).
// Credentials are optional — if absent, notifications are logged and skipped gracefully.
type NotificationService struct {
	sendGridAPIKey string
	fromEmail      string
	fcmServerKey   string
}

// NewNotificationService creates a notification service without credentials (log-only mode).
func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// WithSendGrid sets SendGrid credentials for email notifications.
func (s *NotificationService) WithSendGrid(apiKey, fromEmail string) *NotificationService {
	s.sendGridAPIKey = apiKey
	s.fromEmail = fromEmail
	return s
}

// WithFCM sets FCM server key for push notifications.
func (s *NotificationService) WithFCM(serverKey string) *NotificationService {
	s.fcmServerKey = serverKey
	return s
}

// sendEmail sends a transactional email via SendGrid. Falls back to log if not configured.
func (s *NotificationService) sendEmail(ctx context.Context, toEmail, subject, body string) error {
	if s.sendGridAPIKey == "" {
		logging.Logger.Info("[notification] email (sendgrid not configured)",
			zap.String("to", toEmail),
			zap.String("subject", subject),
		)
		return nil
	}

	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{"to": []map[string]string{{"email": toEmail}}},
		},
		"from":    map[string]string{"email": s.fromEmail},
		"subject": subject,
		"content": []map[string]string{
			{"type": "text/plain", "value": body},
		},
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.sendgrid.com/v3/mail/send", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("sendgrid: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.sendGridAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sendgrid: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// sendPush sends an FCM push notification. Falls back to log if not configured or no token.
func (s *NotificationService) sendPush(ctx context.Context, deviceToken, title, body string) error {
	if s.fcmServerKey == "" || deviceToken == "" {
		logging.Logger.Info("[notification] push (fcm not configured or no token)",
			zap.String("title", title),
		)
		return nil
	}

	payload := map[string]interface{}{
		"to":           deviceToken,
		"notification": map[string]string{"title": title, "body": body},
	}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://fcm.googleapis.com/fcm/send", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("fcm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+s.fcmServerKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("fcm: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("fcm: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// SendGracePeriodExpiringNotification sends a notification when grace period is expiring soon.
func (s *NotificationService) SendGracePeriodExpiringNotification(ctx context.Context, userID uuid.UUID, gracePeriod *entity.GracePeriod) error {
	subject := "Your subscription grace period is expiring soon"
	body := fmt.Sprintf("Your grace period for subscription %s expires in %d hours. Please update your payment method.",
		gracePeriod.SubscriptionID, gracePeriod.HoursRemaining())
	// Email requires user email lookup — log user_id for now, push uses device token from payload
	logging.Logger.Info("grace period expiring notification",
		zap.String("user_id", userID.String()),
		zap.String("subscription_id", gracePeriod.SubscriptionID.String()),
		zap.Int("hours_remaining", gracePeriod.HoursRemaining()),
	)
	_ = subject
	_ = body
	return s.sendPush(ctx, "", subject, body)
}

// SendWinbackOfferNotification sends a winback offer to churned users.
func (s *NotificationService) SendWinbackOfferNotification(ctx context.Context, userID uuid.UUID, offer *entity.WinbackOffer) error {
	title := "We miss you! Special offer inside"
	body := fmt.Sprintf("Come back and save %.0f%% on your subscription.", offer.DiscountValue)
	logging.Logger.Info("winback offer notification",
		zap.String("user_id", userID.String()),
		zap.String("campaign_id", offer.CampaignID),
		zap.Float64("discount", offer.DiscountValue),
	)
	return s.sendPush(ctx, "", title, body)
}

// SendSubscriptionExpiredNotification sends notification when subscription expires.
func (s *NotificationService) SendSubscriptionExpiredNotification(ctx context.Context, userID uuid.UUID, subscriptionID uuid.UUID) error {
	title := "Your subscription has expired"
	body := "Renew now to continue enjoying premium features."
	logging.Logger.Info("subscription expired notification",
		zap.String("user_id", userID.String()),
		zap.String("subscription_id", subscriptionID.String()),
	)
	return s.sendPush(ctx, "", title, body)
}

// SendPaymentRetryNotification sends a notification about failed payment and retry attempt.
func (s *NotificationService) SendPaymentRetryNotification(ctx context.Context, userID uuid.UUID, retryCount int) error {
	title := "Payment failed"
	body := fmt.Sprintf("We could not process your payment (attempt %d). Please update your payment method.", retryCount)
	logging.Logger.Info("payment retry notification",
		zap.String("user_id", userID.String()),
		zap.Int("retry_count", retryCount),
	)
	return s.sendPush(ctx, "", title, body)
}

// SendPaymentSuccessNotification sends a notification when payment is recovered.
func (s *NotificationService) SendPaymentSuccessNotification(ctx context.Context, userID uuid.UUID) {
	title := "Payment successful"
	body := "Your subscription has been renewed successfully."
	logging.Logger.Info("payment success notification",
		zap.String("user_id", userID.String()),
	)
	_ = s.sendPush(ctx, "", title, body)
}

// SendAllRetriesFailedNotification sends a notification when all payment retries fail.
func (s *NotificationService) SendAllRetriesFailedNotification(ctx context.Context, userID uuid.UUID) {
	title := "Subscription cancelled"
	body := "We were unable to process your payment. Your subscription has been cancelled."
	logging.Logger.Info("all retries failed notification",
		zap.String("user_id", userID.String()),
	)
	_ = s.sendPush(ctx, "", title, body)
}

// SendPaymentFinalFailureNotification is an alias for SendAllRetriesFailedNotification.
func (s *NotificationService) SendPaymentFinalFailureNotification(ctx context.Context, userID uuid.UUID) {
	s.SendAllRetriesFailedNotification(ctx, userID)
}
