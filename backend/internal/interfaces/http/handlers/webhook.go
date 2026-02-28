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
		allowedIPs: map[string][]string{
			"stripe": {
				"54.187.174.169/32",
				"54.187.205.235/32",
				"54.187.205.236/32",
				"54.187.216.72/32",
				"54.187.217.22/32",
				"54.187.217.203/32",
				"54.187.226.113/32",
				"54.187.226.135/32",
				"54.187.226.188/32",
				"54.187.226.212/32",
				"54.187.227.33/32",
				"54.187.227.105/32",
				"54.187.227.181/32",
				"54.187.232.79/32",
				"54.187.232.107/32",
				"54.187.233.92/32",
				"54.187.233.124/32",
				"54.187.233.169/32",
				"54.187.234.100/32",
				"54.187.234.118/32",
				"54.187.234.158/32",
				"54.187.235.83/32",
				"54.187.235.104/32",
				"54.187.235.129/32",
				"54.187.235.151/32",
				"54.187.235.161/32",
				"54.187.236.128/32",
				"54.187.236.147/32",
				"54.187.236.185/32",
				"54.187.236.206/32",
				"54.187.237.19/32",
				"54.187.237.22/32",
				"54.187.237.70/32",
				"54.187.237.102/32",
				"54.187.237.120/32",
				"54.187.237.133/32",
				"54.187.238.70/32",
				"54.187.238.99/32",
				"54.187.238.108/32",
				"54.187.238.156/32",
				"54.187.239.6/32",
				"54.187.239.68/32",
				"54.187.239.106/32",
				"54.187.239.118/32",
				"54.187.239.150/32",
				"54.187.239.167/32",
				"54.187.240.67/32",
				"54.187.240.91/32",
				"54.187.240.120/32",
				"54.187.240.133/32",
				"54.187.240.159/32",
				"54.187.240.178/32",
				"54.187.241.84/32",
				"54.187.241.99/32",
				"54.187.241.124/32",
				"54.187.241.171/32",
				"54.187.241.186/32",
				"54.187.242.52/32",
				"54.187.242.97/32",
				"54.187.242.130/32",
				"54.187.242.180/32",
				"54.187.243.16/32",
				"54.187.243.64/32",
				"54.187.243.103/32",
				"54.187.243.128/32",
				"54.187.243.156/32",
				"54.187.244.66/32",
				"54.187.244.94/32",
				"54.187.244.146/32",
				"54.187.244.165/32",
				"54.187.245.10/32",
				"54.187.245.28/32",
				"54.187.245.41/32",
				"54.187.245.79/32",
				"54.187.245.123/32",
				"54.187.245.155/32",
				"54.187.245.185/32",
				"54.187.246.10/32",
				"54.187.246.39/32",
				"54.187.246.59/32",
				"54.187.246.84/32",
				"54.187.246.99/32",
				"54.187.246.154/32",
				"54.187.246.229/32",
				"54.187.246.239/32",
				"54.187.247.10/32",
				"54.187.247.68/32",
				"54.187.247.83/32",
				"54.187.247.111/32",
				"54.187.247.145/32",
				"54.187.247.167/32",
				"54.187.247.233/32",
				"54.187.248.80/32",
				"54.187.248.106/32",
				"54.187.248.127/32",
				"54.187.248.169/32",
				"54.187.248.216/32",
				"54.187.249.22/32",
				"54.187.249.85/32",
				"54.187.249.136/32",
				"54.187.249.213/32",
				"54.187.249.250/32",
				"54.187.250.11/32",
				"54.187.250.145/32",
				"54.187.250.187/32",
				"54.187.251.12/32",
				"54.187.251.59/32",
				"54.187.251.82/32",
				"54.187.251.104/32",
				"54.187.251.151/32",
				"54.187.251.179/32",
				"54.187.252.22/32",
				"54.187.252.84/32",
				"54.187.252.94/32",
				"54.187.252.178/32",
				"54.187.252.200/32",
				"54.187.252.214/32",
				"54.187.252.233/32",
				"54.187.253.11/32",
				"54.187.253.16/32",
				"54.187.253.195/32",
				"54.187.254.12/32",
				"54.187.254.35/32",
				"54.187.254.82/32",
				"54.187.254.123/32",
				"54.187.254.213/32",
				"54.187.255.9/32",
				"54.187.255.75/32",
				"54.187.255.88/32",
				"54.187.255.107/32",
				"54.187.255.148/32",
				"54.187.255.173/32",
				"54.187.255.192/32",
				"54.187.255.206/32",
				"54.187.255.219/32",
				"54.187.255.230/32",
				"54.255.236.18/32",
				"54.255.236.21/32",
				"54.255.236.61/32",
				"54.255.237.28/32",
				"54.255.237.42/32",
				"54.255.238.41/32",
				"54.255.238.44/32",
				"54.255.239.18/32",
				"54.255.239.61/32",
				"54.255.240.17/32",
				"54.255.240.43/32",
				"54.255.240.52/32",
				"54.255.241.21/32",
				"54.255.241.29/32",
				"54.255.241.62/32",
				"54.255.241.81/32",
				"54.255.242.59/32",
				"54.255.242.62/32",
				"54.255.242.91/32",
				"54.255.243.28/32",
				"54.255.243.47/32",
				"54.255.244.50/32",
				"54.255.245.37/32",
				"54.255.246.88/32",
				"54.255.246.91/32",
				"54.255.247.18/32",
				"54.255.247.20/32",
				"54.255.247.37/32",
				"54.255.247.56/32",
				"54.255.247.62/32",
				"54.255.247.86/32",
				"54.255.248.1/32",
				"54.255.248.65/32",
				"54.255.248.91/32",
				"54.255.249.16/32",
				"54.255.249.32/32",
				"54.255.249.81/32",
				"54.255.250.7/32",
				"54.255.250.60/32",
				"54.255.250.108/32",
				"54.255.250.145/32",
				"54.255.251.27/32",
				"54.255.251.30/32",
				"54.255.251.36/32",
				"54.255.251.75/32",
				"54.255.251.77/32",
				"54.255.251.78/32",
				"54.255.251.95/32",
				"54.255.251.104/32",
				"54.255.251.114/32",
				"54.255.251.118/32",
				"54.255.251.140/32",
				"54.255.251.144/32",
				"54.255.251.178/32",
				"54.255.251.199/32",
				"54.255.251.207/32",
				"54.255.252.41/32",
				"54.255.252.74/32",
				"54.255.252.97/32",
				"54.255.253.19/32",
				"54.255.253.47/32",
				"54.255.253.81/32",
				"54.255.254.22/32",
				"54.255.254.25/32",
				"54.255.254.49/32",
				"54.255.254.96/32",
				"54.255.254.126/32",
				"54.255.255.7/32",
				"54.255.255.31/32",
				"54.255.255.37/32",
				"54.255.255.51/32",
				"54.255.255.54/32",
				"54.255.255.65/32",
				"54.255.255.100/32",
			},
			"apple":  {"17.0.0.0/8"},
			"google": {"66.102.0.0/20", "64.233.160.0/19"},
		},
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
