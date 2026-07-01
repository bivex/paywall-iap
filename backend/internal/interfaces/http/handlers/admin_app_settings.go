package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bivex/paywall-iap/internal/domain/entity"
	domainErrors "github.com/bivex/paywall-iap/internal/domain/errors"
	domainRepo "github.com/bivex/paywall-iap/internal/domain/repository"
	iapext "github.com/bivex/paywall-iap/internal/infrastructure/external/iap"
	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

// AppSettingsHandler handles /v1/admin/apps/:id/settings and /v1/admin/apps/:id/credentials
type AppSettingsHandler struct {
	appRepo  domainRepo.AppRepository
	resolver *iapext.CredentialResolver
}

func NewAppSettingsHandler(appRepo domainRepo.AppRepository, resolver *iapext.CredentialResolver) *AppSettingsHandler {
	return &AppSettingsHandler{appRepo: appRepo, resolver: resolver}
}

// ── Settings ──────────────────────────────────────────────────────────────────

type appSettingsRequest struct {
	GracePeriodDays         *int                `json:"grace_period_days"`
	TrialEnabled            *bool               `json:"trial_enabled"`
	TrialDays               *int                `json:"trial_days"`
	DefaultCurrency         *string             `json:"default_currency"`
	WebhookURL              *string             `json:"webhook_url"`
	WebhookSecret           *string             `json:"webhook_secret"`
	StoreEnvironment        *string             `json:"store_environment" binding:"omitempty,oneof=production sandbox"`
	Entitlements            map[string][]string `json:"entitlements"`
	SubscriptionRequiredFor []string            `json:"subscription_required_for"`
}

// GetAppSettings GET /v1/admin/apps/:id/settings
func (h *AppSettingsHandler) GetAppSettings(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid app id")
		return
	}
	s, err := h.appRepo.GetSettings(c.Request.Context(), id)
	if err != nil {
		if isNotFound(err) {
			response.NotFound(c, "app not found")
			return
		}
		response.InternalError(c, "failed to get app settings")
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": s})
}

// PutAppSettings PUT /v1/admin/apps/:id/settings
func (h *AppSettingsHandler) PutAppSettings(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid app id")
		return
	}
	var req appSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Load current settings so we do a merge, not a blind overwrite.
	current, err := h.appRepo.GetSettings(c.Request.Context(), id)
	if err != nil {
		if isNotFound(err) {
			response.NotFound(c, "app not found")
			return
		}
		response.InternalError(c, "failed to load app settings")
		return
	}

	if req.GracePeriodDays != nil {
		if *req.GracePeriodDays < 0 || *req.GracePeriodDays > 90 {
			response.UnprocessableEntity(c, "grace_period_days must be 0–90")
			return
		}
		current.GracePeriodDays = *req.GracePeriodDays
	}
	if req.TrialEnabled != nil {
		current.TrialEnabled = *req.TrialEnabled
	}
	if req.TrialDays != nil {
		if *req.TrialDays < 0 || *req.TrialDays > 365 {
			response.UnprocessableEntity(c, "trial_days must be 0–365")
			return
		}
		current.TrialDays = *req.TrialDays
	}
	if req.DefaultCurrency != nil {
		cur := strings.ToUpper(strings.TrimSpace(*req.DefaultCurrency))
		if len(cur) != 3 {
			response.UnprocessableEntity(c, "default_currency must be a 3-letter ISO-4217 code")
			return
		}
		current.DefaultCurrency = cur
	}
	if req.WebhookURL != nil {
		current.WebhookURL = *req.WebhookURL
	}
	if req.WebhookSecret != nil {
		current.WebhookSecret = *req.WebhookSecret
	}
	if req.StoreEnvironment != nil {
		current.StoreEnvironment = *req.StoreEnvironment
	}
	if req.Entitlements != nil {
		current.Entitlements = req.Entitlements
	}
	if req.SubscriptionRequiredFor != nil {
		current.SubscriptionRequiredFor = req.SubscriptionRequiredFor
	}

	if err := h.appRepo.UpdateSettings(c.Request.Context(), id, current); err != nil {
		if isNotFound(err) {
			response.NotFound(c, "app not found")
			return
		}
		response.InternalError(c, "failed to update app settings")
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": current})
}

// ── Credentials ───────────────────────────────────────────────────────────────

type credentialsRequest struct {
	Provider string `json:"provider" binding:"required,oneof=apple google stripe paddle"`

	// Apple
	AppleSharedSecret string `json:"apple_shared_secret"`
	AppleTeamID       string `json:"apple_team_id"`
	AppleKeyID        string `json:"apple_key_id"`
	ApplePrivateKey   string `json:"apple_private_key"`
	AppleBundleID     string `json:"apple_bundle_id"`
	AppleEnvironment  string `json:"apple_environment" binding:"omitempty,oneof=production sandbox"`

	// Google
	GooglePackageName    string `json:"google_package_name"`
	GoogleServiceAccount string `json:"google_service_account"`

	// Stripe
	StripePublishableKey string `json:"stripe_publishable_key"`
	StripeSecretKey      string `json:"stripe_secret_key"`
	StripeWebhookSecret  string `json:"stripe_webhook_secret"`

	// Paddle
	PaddleVendorID      string `json:"paddle_vendor_id"`
	PaddleAPIKey        string `json:"paddle_api_key"`
	PaddleWebhookSecret string `json:"paddle_webhook_secret"`
}

// credentialsDTO is what we return — sensitive fields masked.
type credentialsDTO struct {
	Provider string `json:"provider"`

	AppleTeamID      string `json:"apple_team_id,omitempty"`
	AppleKeyID       string `json:"apple_key_id,omitempty"`
	AppleBundleID    string `json:"apple_bundle_id,omitempty"`
	AppleEnvironment string `json:"apple_environment,omitempty"`
	// encrypted fields shown as boolean "configured"
	AppleSharedSecretSet bool `json:"apple_shared_secret_set"`
	ApplePrivateKeySet   bool `json:"apple_private_key_set"`

	GooglePackageName    string `json:"google_package_name,omitempty"`
	GoogleServiceAccountSet bool `json:"google_service_account_set"`

	StripePublishableKey  string `json:"stripe_publishable_key,omitempty"`
	StripeSecretKeySet    bool   `json:"stripe_secret_key_set"`
	StripeWebhookSecretSet bool  `json:"stripe_webhook_secret_set"`

	PaddleVendorID          string `json:"paddle_vendor_id,omitempty"`
	PaddleAPIKeySet         bool   `json:"paddle_api_key_set"`
	PaddleWebhookSecretSet  bool   `json:"paddle_webhook_secret_set"`
}

func toCredentialsDTO(c *entity.AppCredentials) credentialsDTO {
	return credentialsDTO{
		Provider:                c.Provider,
		AppleTeamID:             c.AppleTeamID,
		AppleKeyID:              c.AppleKeyID,
		AppleBundleID:           c.AppleBundleID,
		AppleEnvironment:        c.AppleEnvironment,
		AppleSharedSecretSet:    c.AppleSharedSecret != "",
		ApplePrivateKeySet:      c.ApplePrivateKey != "",
		GooglePackageName:       c.GooglePackageName,
		GoogleServiceAccountSet: c.GoogleServiceAccount != "",
		StripePublishableKey:    c.StripePublishableKey,
		StripeSecretKeySet:      c.StripeSecretKey != "",
		StripeWebhookSecretSet:  c.StripeWebhookSecret != "",
		PaddleVendorID:          c.PaddleVendorID,
		PaddleAPIKeySet:         c.PaddleAPIKey != "",
		PaddleWebhookSecretSet:  c.PaddleWebhookSecret != "",
	}
}

// GetAppCredentials GET /v1/admin/apps/:id/credentials
func (h *AppSettingsHandler) GetAppCredentials(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid app id")
		return
	}
	creds, err := h.appRepo.GetCredentials(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to get credentials")
		return
	}
	dtos := make([]credentialsDTO, 0, len(creds))
	for _, cr := range creds {
		dtos = append(dtos, toCredentialsDTO(cr))
	}
	c.JSON(http.StatusOK, gin.H{"credentials": dtos})
}

// PutAppCredentials PUT /v1/admin/apps/:id/credentials
func (h *AppSettingsHandler) PutAppCredentials(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid app id")
		return
	}
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	appleEnv := req.AppleEnvironment
	if appleEnv == "" {
		appleEnv = "production"
	}

	creds := &entity.AppCredentials{
		AppID:                id,
		Provider:             req.Provider,
		AppleSharedSecret:    req.AppleSharedSecret,
		AppleTeamID:          req.AppleTeamID,
		AppleKeyID:           req.AppleKeyID,
		ApplePrivateKey:      req.ApplePrivateKey,
		AppleBundleID:        req.AppleBundleID,
		AppleEnvironment:     appleEnv,
		GooglePackageName:    req.GooglePackageName,
		GoogleServiceAccount: req.GoogleServiceAccount,
		StripePublishableKey: req.StripePublishableKey,
		StripeSecretKey:      req.StripeSecretKey,
		StripeWebhookSecret:  req.StripeWebhookSecret,
		PaddleVendorID:       req.PaddleVendorID,
		PaddleAPIKey:         req.PaddleAPIKey,
		PaddleWebhookSecret:  req.PaddleWebhookSecret,
	}

	if err := h.appRepo.UpsertCredentials(c.Request.Context(), creds); err != nil {
		response.InternalError(c, "failed to save credentials")
		return
	}

	// Invalidate credential cache so the next IAP verify picks up the new keys immediately.
	if h.resolver != nil {
		h.resolver.Invalidate(id, creds.Provider)
	}

	c.JSON(http.StatusOK, gin.H{"credentials": toCredentialsDTO(creds)})
}

// DeleteAppCredentials DELETE /v1/admin/apps/:id/credentials/:provider
func (h *AppSettingsHandler) DeleteAppCredentials(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid app id")
		return
	}
	provider := c.Param("provider")
	valid := map[string]bool{"apple": true, "google": true, "stripe": true, "paddle": true}
	if !valid[provider] {
		response.BadRequest(c, "provider must be apple, google, stripe, or paddle")
		return
	}
	if err := h.appRepo.DeleteCredentials(c.Request.Context(), id, provider); err != nil {
		response.InternalError(c, "failed to delete credentials")
		return
	}
	// Invalidate credential cache so stale keys are not used after deletion.
	if h.resolver != nil {
		h.resolver.Invalidate(id, provider)
	}
	response.NoContent(c)
}

// isNotFound checks if error wraps domain ErrNotFound.
func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), domainErrors.ErrNotFound.Error())
}
