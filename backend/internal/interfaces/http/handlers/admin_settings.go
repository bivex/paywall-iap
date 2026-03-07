package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/mail"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/bivex/paywall-iap/internal/interfaces/http/response"
)

const platformSettingsKey = "platform_settings"

type PlatformSettings struct {
	General       GeneralSettings       `json:"general"`
	Integrations  IntegrationSettings   `json:"integrations"`
	Notifications NotificationSettings  `json:"notifications"`
	Security      PlatformSecurityState `json:"security"`
}

type GeneralSettings struct {
	PlatformName    string `json:"platform_name"`
	SupportEmail    string `json:"support_email"`
	DefaultCurrency string `json:"default_currency"`
	DarkModeDefault bool   `json:"dark_mode_default"`
}

type IntegrationSettings struct {
	StripeAPIKey         string `json:"stripe_api_key"`
	StripeWebhookSecret  string `json:"stripe_webhook_secret"`
	StripeTestMode       bool   `json:"stripe_test_mode"`
	AppleIssuerID        string `json:"apple_issuer_id"`
	AppleBundleID        string `json:"apple_bundle_id"`
	GoogleServiceAccount string `json:"google_service_account"`
	GooglePackageName    string `json:"google_package_name"`
	MatomoURL            string `json:"matomo_url"`
	MatomoSiteID         string `json:"matomo_site_id"`
	MatomoAuthToken      string `json:"matomo_auth_token"`
}

type NotificationSettings struct {
	NewSubscription       bool `json:"new_subscription"`
	PaymentFailed         bool `json:"payment_failed"`
	SubscriptionCancelled bool `json:"subscription_cancelled"`
	RefundIssued          bool `json:"refund_issued"`
	WebhookFailed         bool `json:"webhook_failed"`
	DunningStarted        bool `json:"dunning_started"`
}

type PlatformSecurityState struct {
	JWTExpiryHours    int  `json:"jwt_expiry_hours"`
	RequireMFA        bool `json:"require_mfa"`
	EnableIPAllowlist bool `json:"enable_ip_allowlist"`
}

func defaultPlatformSettings() PlatformSettings {
	return PlatformSettings{
		General: GeneralSettings{
			PlatformName:    "Paywall SaaS",
			SupportEmail:    "support@paywall.local",
			DefaultCurrency: "USD",
		},
		Notifications: NotificationSettings{
			NewSubscription:       true,
			PaymentFailed:         true,
			SubscriptionCancelled: true,
			RefundIssued:          true,
			WebhookFailed:         true,
			DunningStarted:        true,
		},
		Security: PlatformSecurityState{
			JWTExpiryHours: 24,
		},
	}
}

func normalizePlatformSettings(settings PlatformSettings) PlatformSettings {
	settings.General.PlatformName = strings.TrimSpace(settings.General.PlatformName)
	settings.General.SupportEmail = strings.TrimSpace(settings.General.SupportEmail)
	settings.General.DefaultCurrency = strings.ToUpper(strings.TrimSpace(settings.General.DefaultCurrency))

	settings.Integrations.StripeAPIKey = strings.TrimSpace(settings.Integrations.StripeAPIKey)
	settings.Integrations.StripeWebhookSecret = strings.TrimSpace(settings.Integrations.StripeWebhookSecret)
	settings.Integrations.AppleIssuerID = strings.TrimSpace(settings.Integrations.AppleIssuerID)
	settings.Integrations.AppleBundleID = strings.TrimSpace(settings.Integrations.AppleBundleID)
	settings.Integrations.GoogleServiceAccount = strings.TrimSpace(settings.Integrations.GoogleServiceAccount)
	settings.Integrations.GooglePackageName = strings.TrimSpace(settings.Integrations.GooglePackageName)
	settings.Integrations.MatomoURL = strings.TrimSpace(settings.Integrations.MatomoURL)
	settings.Integrations.MatomoSiteID = strings.TrimSpace(settings.Integrations.MatomoSiteID)
	settings.Integrations.MatomoAuthToken = strings.TrimSpace(settings.Integrations.MatomoAuthToken)

	return settings
}

func validatePlatformSettings(settings PlatformSettings) string {
	if settings.General.PlatformName == "" {
		return "Platform name is required"
	}
	if _, err := mail.ParseAddress(settings.General.SupportEmail); err != nil {
		return "Support email must be a valid email address"
	}
	if len(settings.General.DefaultCurrency) != 3 {
		return "Default currency must be a 3-letter ISO code"
	}
	if settings.Security.JWTExpiryHours < 1 || settings.Security.JWTExpiryHours > 720 {
		return "JWT expiry must be between 1 and 720 hours"
	}
	if settings.Integrations.MatomoURL != "" {
		u, err := url.ParseRequestURI(settings.Integrations.MatomoURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return "Matomo URL must be a valid absolute URL"
		}
	}
	return ""
}

func (h *AdminHandler) loadPlatformSettings(ctx context.Context) (PlatformSettings, error) {
	settings := defaultPlatformSettings()
	var raw []byte
	err := h.dbPool.QueryRow(ctx, `SELECT value FROM admin_settings WHERE key = $1`, platformSettingsKey).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return settings, nil
		}
		return PlatformSettings{}, err
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return PlatformSettings{}, err
	}
	return settings, nil
}

func (h *AdminHandler) savePlatformSettings(ctx context.Context, settings PlatformSettings) error {
	payload, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	_, err = h.dbPool.Exec(
		ctx,
		`INSERT INTO admin_settings (key, value, updated_at)
		 VALUES ($1, $2::jsonb, now())
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`,
		platformSettingsKey,
		payload,
	)
	return err
}

func (h *AdminHandler) GetPlatformSettings(c *gin.Context) {
	settings, err := h.loadPlatformSettings(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load platform settings")
		return
	}
	response.OK(c, settings)
}

func (h *AdminHandler) UpdatePlatformSettings(c *gin.Context) {
	var req PlatformSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid settings payload")
		return
	}

	settings := normalizePlatformSettings(req)
	if msg := validatePlatformSettings(settings); msg != "" {
		response.UnprocessableEntity(c, msg)
		return
	}
	ctx := c.Request.Context()
	if err := h.savePlatformSettings(ctx, settings); err != nil {
		response.InternalError(c, "Failed to save platform settings")
		return
	}

	adminID, _ := c.Get("admin_id")
	if aid, ok := adminID.(uuid.UUID); ok {
		_ = h.auditService.LogAction(ctx, aid, "update_platform_settings", "admin_settings", &aid, map[string]interface{}{
			"support_email":       settings.General.SupportEmail,
			"default_currency":    settings.General.DefaultCurrency,
			"jwt_expiry_hours":    settings.Security.JWTExpiryHours,
			"require_mfa":         settings.Security.RequireMFA,
			"enable_ip_allowlist": settings.Security.EnableIPAllowlist,
		})
	}

	response.OK(c, settings)
}

func (h *AdminHandler) ChangeAdminPassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid password payload")
		return
	}

	req.CurrentPassword = strings.TrimSpace(req.CurrentPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	req.ConfirmPassword = strings.TrimSpace(req.ConfirmPassword)

	if req.CurrentPassword == "" || req.NewPassword == "" || req.ConfirmPassword == "" {
		response.UnprocessableEntity(c, "All password fields are required")
		return
	}
	if len(req.NewPassword) < 8 {
		response.UnprocessableEntity(c, "New password must be at least 8 characters long")
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		response.UnprocessableEntity(c, "Password confirmation does not match")
		return
	}

	adminIDValue, ok := c.Get("admin_id")
	if !ok {
		response.Unauthorized(c, "Admin context is missing")
		return
	}
	adminID, ok := adminIDValue.(uuid.UUID)
	if !ok {
		response.Unauthorized(c, "Invalid admin context")
		return
	}

	ctx := c.Request.Context()
	cred, err := h.queries.GetAdminCredential(ctx, adminID)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(req.CurrentPassword)) != nil {
		response.UnprocessableEntity(c, "Current password is incorrect")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(req.NewPassword)) == nil {
		response.UnprocessableEntity(c, "New password must be different from the current password")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "Failed to update password")
		return
	}
	if _, err := h.queries.UpsertAdminCredential(ctx, adminID, string(hash)); err != nil {
		response.InternalError(c, "Failed to update password")
		return
	}

	_ = h.auditService.LogAction(ctx, adminID, "change_admin_password", "admin_settings", &adminID, map[string]interface{}{})
	response.OK(c, gin.H{"ok": true})
}
