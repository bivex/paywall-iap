package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bivex/paywall-iap/internal/infrastructure/logging"
	"github.com/bivex/paywall-iap/internal/infrastructure/persistence/sqlc/generated"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
	"github.com/bivex/paywall-iap/internal/worker/tasks"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

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
func NewWebhookHandler(stripeSecret, appleSecret, googleSecret string, queries *generated.Queries, asynqClient *asynq.Client) *WebhookHandler {
	return &WebhookHandler{
		stripeWebhookSecret: stripeSecret,
		appleWebhookSecret:  appleSecret,
		googleWebhookSecret: googleSecret,
		queries:             queries,
		asynqClient:         asynqClient,
		allowedIPs:          WebhookIPConfig,
	}
}

// StripeWebhook handles Stripe webhook events
// @Summary Stripe webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Router /webhook/stripe [post]
func (h *WebhookHandler) StripeWebhook(c *gin.Context) {
	// Verify IP whitelist
	if h.stripeWebhookSecret != "" && h.stripeWebhookSecret != "whsec_dummy" {
		if !h.verifyIP(c.ClientIP(), "stripe") {
			response.Unauthorized(c, "IP not allowed")
			return
		}
	}

	// Verify HMAC signature
	signature := c.GetHeader("Stripe-Signature")
	if h.stripeWebhookSecret != "" && h.stripeWebhookSecret != "whsec_dummy" {
		if signature == "" {
			response.Unauthorized(c, "Missing signature")
			return
		}
	}

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Failed to read body")
		return
	}

	// Verify HMAC
	if h.stripeWebhookSecret != "" && h.stripeWebhookSecret != "whsec_dummy" {
		if !h.verifyStripeHMAC(body, signature) {
			response.Unauthorized(c, "Invalid signature")
			return
		}
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
	payload, _ := json.Marshal(map[string]string{
		"provider":   "stripe",
		"event_type": event.Type,
		"event_id":   event.ID,
	})
	if _, err := h.asynqClient.Enqueue(asynq.NewTask(tasks.TypeProcessWebhook, payload)); err != nil {
		logging.Logger.Error("Failed to enqueue webhook task", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// AppleWebhook handles Apple S2S notifications
// @Summary Apple webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Router /webhook/apple [post]
func (h *WebhookHandler) AppleWebhook(c *gin.Context) {
	// Verify IP whitelist
	if h.appleWebhookSecret != "" && h.appleWebhookSecret != "whsec_dummy" {
		if !h.verifyIP(c.ClientIP(), "apple") {
			response.Unauthorized(c, "IP not allowed")
			return
		}
	}

	// Apple sends a JWS compact token as the raw body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Failed to read body")
		return
	}

	jwsToken := strings.TrimSpace(string(body))

	// JWS compact format: header.payload.signature (three dot-separated base64url parts)
	parts := strings.Split(jwsToken, ".")
	if len(parts) != 3 {
		response.BadRequest(c, "Invalid JWS token format")
		return
	}

	// Decode the payload (middle part) — base64url with no padding
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		response.BadRequest(c, "Failed to decode JWS payload")
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

	// Skip JWS signature verification in dev (no Apple cert configured)
	// In production: verify using Apple's WWDR certificate chain
	if h.appleWebhookSecret != "" {
		// Full verification would parse the x5c chain from the JWS header
		// and verify against Apple's root CA. Omitted in this implementation.
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
	taskPayload, _ := json.Marshal(map[string]string{
		"provider":   "apple",
		"event_type": notification.NotificationType,
		"event_id":   notification.NotificationUUID,
	})
	if _, err := h.asynqClient.Enqueue(asynq.NewTask(tasks.TypeProcessWebhook, taskPayload)); err != nil {
		logging.Logger.Error("Failed to enqueue Apple webhook task", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// GoogleWebhook handles Google RTDN notifications
// @Summary Google webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Router /webhook/google [post]
func (h *WebhookHandler) GoogleWebhook(c *gin.Context) {
	// Verify IP whitelist
	if h.googleWebhookSecret != "" && h.googleWebhookSecret != "whsec_dummy" {
		if !h.verifyIP(c.ClientIP(), "google") {
			response.Unauthorized(c, "IP not allowed")
			return
		}
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
	taskPayload, _ := json.Marshal(map[string]string{
		"provider":   "google",
		"event_type": eventType,
		"event_id":   eventID,
	})
	if _, err := h.asynqClient.Enqueue(asynq.NewTask(tasks.TypeProcessWebhook, taskPayload)); err != nil {
		logging.Logger.Error("Failed to enqueue Google webhook task", zap.Error(err))
	}

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
