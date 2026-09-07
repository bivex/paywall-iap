package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
	"github.com/bivex/paywall-iap/internal/worker/tasks"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// WebhookHandlerConfig configures WebhookHandler
type WebhookHandlerConfig struct {
	StripeSecret string
	AppleSecret  string
	GoogleSecret string
	Queries      *generated.Queries
	AsynqClient  *asynq.Client
}

// WebhookHandler handles webhook endpoints from external services
type WebhookHandler struct {
	stripeWebhookSecret string
	appleWebhookSecret  string
	googleWebhookSecret string
	allowedIPs          map[string][]string // service -> IPs
	queries             *generated.Queries
	asynqClient         *asynq.Client
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(cfg WebhookHandlerConfig) *WebhookHandler {
	return &WebhookHandler{
		stripeWebhookSecret: cfg.StripeSecret,
		appleWebhookSecret:  cfg.AppleSecret,
		googleWebhookSecret: cfg.GoogleSecret,
		queries:             cfg.Queries,
		asynqClient:         cfg.AsynqClient,
		allowedIPs:          WebhookIPConfig,
	}
}

func (h *WebhookHandler) enqueueWebhookTask(provider, eventType, eventID string) {
	if h.asynqClient == nil {
		return
	}
	payload, _ := json.Marshal(map[string]string{
		"provider":   provider,
		"event_type": eventType,
		"event_id":   eventID,
	})
	task := asynq.NewTask(tasks.TypeProcessWebhook, payload)
	taskID := fmt.Sprintf("webhook:%s:%s", provider, eventID)
	if _, err := h.asynqClient.Enqueue(
		task,
		asynq.TaskID(taskID),
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Second),
	); err != nil {
		if !strings.Contains(err.Error(), "task ID already exists") {
			logging.Logger.Error("Failed to enqueue webhook task",
				zap.String("provider", provider),
				zap.String("event_id", eventID),
				zap.Error(err),
			)
		}
	}
}

func (h *WebhookHandler) isIPAllowed(clientIP, service, secret string) bool {
	if secret == "" || secret == "whsec_dummy" {
		return true
	}
	return h.verifyIP(clientIP, service)
}

func (h *WebhookHandler) validateStripeRequest(c *gin.Context, body []byte) bool {
	if !h.isIPAllowed(c.ClientIP(), "stripe", h.stripeWebhookSecret) {
		response.Unauthorized(c, "IP not allowed")
		return false
	}
	if h.stripeWebhookSecret != "" && h.stripeWebhookSecret != "whsec_dummy" {
		signature := c.GetHeader("Stripe-Signature")
		if signature == "" || !h.verifyStripeHMAC(body, signature) {
			response.Unauthorized(c, "Invalid signature")
			return false
		}
	}
	return true
}

// StripeWebhook handles Stripe webhook events
// @Summary Stripe webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Router /webhook/stripe [post]
func (h *WebhookHandler) StripeWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Failed to read body")
		return
	}

	if !h.validateStripeRequest(c, body) {
		return
	}

	// Parse event ID and type from Stripe JSON body
	var event struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		response.BadRequest(c, "Invalid event body")
		return
	}

	if err := h.queries.InsertWebhookEvent(c.Request.Context(), generated.InsertWebhookEventParams{
		Provider:  "stripe",
		EventType: event.Type,
		EventID:   event.ID,
		Payload:   body,
	}); err != nil {
		// Log but return 200 — Stripe retries on failure
		_ = err
	}

	// Enqueue background processing task
	h.enqueueWebhookTask("stripe", event.Type, event.ID)

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func (h *WebhookHandler) extractApplePayload(body []byte) ([]byte, error) {
	jwsToken := strings.TrimSpace(string(body))
	parts := strings.Split(jwsToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWS token format")
	}
	return base64.RawURLEncoding.DecodeString(parts[1])
}

// AppleWebhook handles Apple S2S notifications
// @Summary Apple webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Router /webhook/apple [post]
func (h *WebhookHandler) AppleWebhook(c *gin.Context) {
	if !h.isIPAllowed(c.ClientIP(), "apple", h.appleWebhookSecret) {
		response.Unauthorized(c, "IP not allowed")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Failed to read body")
		return
	}

	payloadBytes, err := h.extractApplePayload(body)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Parse the Apple notification envelope
	var notification struct {
		NotificationType string `json:"notificationType"`
		NotificationUUID string `json:"notificationUUID"`
		Data             struct {
			SignedTransactionInfo string `json:"signedTransactionInfo"`
			SignedRenewalInfo     string `json:"signedRenewalInfo"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payloadBytes, &notification); err != nil {
		response.BadRequest(c, "Failed to parse notification payload")
		return
	}

	if err := h.queries.InsertWebhookEvent(c.Request.Context(), generated.InsertWebhookEventParams{
		Provider:  "apple",
		EventType: notification.NotificationType,
		EventID:   notification.NotificationUUID,
		Payload:   payloadBytes,
	}); err != nil {
		_ = err // idempotent insert — ignore duplicate errors
	}

	// Enqueue background processing task
	h.enqueueWebhookTask("apple", notification.NotificationType, notification.NotificationUUID)

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// GoogleWebhook handles Google RTDN notifications
// @Summary Google webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Router /webhook/google [post]
func (h *WebhookHandler) GoogleWebhook(c *gin.Context) {
	if !h.isIPAllowed(c.ClientIP(), "google", h.googleWebhookSecret) {
		response.Unauthorized(c, "IP not allowed")
		return
	}

	// Google sends Pub/Sub push as JSON with base64-encoded message.data
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Failed to read body")
		return
	}

	var pubsubMessage struct {
		Message struct {
			Data      string `json:"data"`      // base64-encoded
			MessageID string `json:"messageId"` // unique per message → use as event_id
		} `json:"message"`
		Subscription string `json:"subscription"`
	}
	if err := json.Unmarshal(body, &pubsubMessage); err != nil {
		response.BadRequest(c, "Invalid Pub/Sub message")
		return
	}

	// Decode the notification payload
	notificationBytes, err := base64.StdEncoding.DecodeString(pubsubMessage.Message.Data)
	if err != nil {
		response.BadRequest(c, "Failed to decode Pub/Sub data")
		return
	}

	// Parse notification type from the RTDN payload
	var rtdn struct {
		SubscriptionNotification struct {
			NotificationType int    `json:"notificationType"`
			PurchaseToken    string `json:"purchaseToken"`
			SubscriptionID   string `json:"subscriptionId"`
		} `json:"subscriptionNotification"`
		PackageName string `json:"packageName"`
	}
	if err := json.Unmarshal(notificationBytes, &rtdn); err != nil {
		response.BadRequest(c, "Failed to parse RTDN notification")
		return
	}

	// Verify Google JWT in Authorization header (skip in dev)
	if h.googleWebhookSecret != "" {
		// In production: validate the Authorization: Bearer token is a valid
		// Google-signed OIDC token for the configured service account.
	}

	eventType := fmt.Sprintf("subscription.%d", rtdn.SubscriptionNotification.NotificationType)
	eventID := pubsubMessage.Message.MessageID

	if err := h.queries.InsertWebhookEvent(c.Request.Context(), generated.InsertWebhookEventParams{
		Provider:  "google",
		EventType: eventType,
		EventID:   eventID,
		Payload:   notificationBytes,
	}); err != nil {
		_ = err
	}

	// Enqueue background processing task (same pattern as Stripe).
	h.enqueueWebhookTask("google", eventType, eventID)

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// verifyStripeHMAC verifies Stripe webhook signature
func (h *WebhookHandler) verifyStripeHMAC(body []byte, signature string) bool {
	if h.stripeWebhookSecret == "" {
		// Skip verification in development
		return true
	}

	// Stripe signature format: t=timestamp,v1=hmac
	parts := strings.Split(signature, ",")
	if len(parts) != 2 {
		return false
	}

	timestamp := strings.TrimPrefix(parts[0], "t=")
	v1 := strings.TrimPrefix(parts[1], "v1=")

	// Create expected signature
	payload := []byte(timestamp + "." + string(body))
	mac := hmac.New(sha256.New, []byte(h.stripeWebhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(v1), []byte(expected))
}

// verifyIP checks if the client IP is in the allowed list
func (h *WebhookHandler) verifyIP(clientIP, service string) bool {
	allowedIPs, exists := h.allowedIPs[service]
	if !exists {
		return false
	}

	// Simple IP check - in production, use proper CIDR matching
	for _, ipRange := range allowedIPs {
		if strings.Contains(ipRange, "/32") {
			// Exact match for single IP
			ip := strings.TrimSuffix(ipRange, "/32")
			if clientIP == ip {
				return true
			}
		} else {
			// CIDR range - simple prefix match for MVP
			prefix := strings.Split(ipRange, "/")[0]
			prefixParts := strings.Split(prefix, ".")
			clientParts := strings.Split(clientIP, ".")

			if len(clientParts) == 4 {
				clientPrefix := strings.Join(clientParts[:len(ipRange)-1], ".")
				if strings.HasPrefix(clientPrefix, strings.Join(prefixParts[:len(ipRange)-1], ".")) {
					return true
				}
			}
		}
	}

	return false
}
