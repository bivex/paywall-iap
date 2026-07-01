package entity

import (
	"time"

	"github.com/google/uuid"
)

// App represents a registered Mothsalt application.
type App struct {
	ID          uuid.UUID
	Name        string // reverse-dns, e.g. "com.mothsalt.game1"
	DisplayName string
	Platform    string // "ios", "android", "both"
	BundleID    string // App Store bundle ID / Google Play package name
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AppSettings holds non-sensitive per-app paywall configuration stored as JSONB.
type AppSettings struct {
	GracePeriodDays          int               `json:"grace_period_days"`
	TrialEnabled             bool              `json:"trial_enabled"`
	TrialDays                int               `json:"trial_days"`
	DefaultCurrency          string            `json:"default_currency"`
	WebhookURL               string            `json:"webhook_url"`
	WebhookSecret            string            `json:"webhook_secret"`
	StoreEnvironment         string            `json:"store_environment"` // "production" | "sandbox"
	Entitlements             map[string][]string `json:"entitlements"`     // product_id → []feature_key
	SubscriptionRequiredFor  []string          `json:"subscription_required_for"`
}

// AppCredentials holds store keys for one provider. Sensitive fields are encrypted at rest.
type AppCredentials struct {
	ID     uuid.UUID
	AppID  uuid.UUID
	Provider string // "apple" | "google" | "stripe" | "paddle"

	// Apple
	AppleSharedSecret string // decrypted at read time
	AppleTeamID       string
	AppleKeyID        string
	ApplePrivateKey   string // decrypted at read time
	AppleBundleID     string
	AppleEnvironment  string // "production" | "sandbox"

	// Google
	GooglePackageName   string
	GoogleServiceAccount string // decrypted at read time

	// Stripe
	StripePublishableKey  string
	StripeSecretKey       string // decrypted
	StripeWebhookSecret   string // decrypted

	// Paddle
	PaddleVendorID      string
	PaddleAPIKey        string // decrypted
	PaddleWebhookSecret string // decrypted

	CreatedAt time.Time
	UpdatedAt time.Time
}
