-- Migration 036: app_settings JSONB + app_credentials (encrypted store keys)
--
-- app_settings  — non-sensitive per-app config (grace period, trial, webhook URL, etc.)
-- app_credentials — sensitive store keys, encrypted at rest via pgcrypto AES-256
--
-- Encryption key is injected at runtime via app.settings.encryption_key (env var APP_CREDENTIALS_KEY).
-- All encrypt/decrypt happens in the application layer; this migration only creates the schema.

-- ── 1. app_settings column on apps ────────────────────────────────────────────

ALTER TABLE apps
    ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN apps.settings IS
'Non-sensitive per-app paywall configuration. Schema:
{
  "grace_period_days":       int,      -- days of access after subscription expires (default 3)
  "trial_enabled":           bool,     -- whether free trial is offered
  "trial_days":              int,      -- trial duration in days
  "default_currency":        "USD",    -- ISO-4217, used when pricing tier has no explicit currency
  "webhook_url":             "https://...",
  "webhook_secret":          "whsec_...",  -- HMAC secret for webhook signature validation
  "entitlements": {                    -- product_id → list of feature keys granted
    "com.example.premium_monthly": ["premium", "no_ads"],
    "com.example.premium_annual":  ["premium", "no_ads", "offline"]
  },
  "store_environment":       "production" | "sandbox",
  "subscription_required_for": ["feature_a", "feature_b"]
}';

-- Validate grace_period_days range when present
ALTER TABLE apps
    ADD CONSTRAINT apps_settings_grace_period_check
    CHECK (
        (settings->>'grace_period_days') IS NULL
        OR (settings->>'grace_period_days')::int BETWEEN 0 AND 90
    );

-- Validate trial_days range when present
ALTER TABLE apps
    ADD CONSTRAINT apps_settings_trial_days_check
    CHECK (
        (settings->>'trial_days') IS NULL
        OR (settings->>'trial_days')::int BETWEEN 0 AND 365
    );

-- ── 2. app_credentials table ───────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS app_credentials (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id       UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,

    -- which store this credential is for
    provider     TEXT NOT NULL CHECK (provider IN ('apple', 'google', 'stripe', 'paddle')),

    -- encrypted fields (AES-256-CBC via application layer, stored as base64 ciphertext)
    -- NULL means not configured for this provider
    apple_shared_secret_enc      TEXT,   -- App-specific shared secret (non-subscription receipt validation)
    apple_team_id                TEXT,   -- 10-char Team ID (not secret, stored plaintext)
    apple_key_id                 TEXT,   -- 10-char Key ID for App Store Server API (not secret)
    apple_private_key_enc        TEXT,   -- .p8 private key contents, encrypted
    apple_bundle_id              TEXT,   -- explicit bundle_id override (falls back to apps.bundle_id)
    apple_environment            TEXT NOT NULL DEFAULT 'production'
                                     CHECK (apple_environment IN ('production', 'sandbox')),

    google_package_name          TEXT,   -- com.example.app (falls back to apps.bundle_id)
    google_service_account_enc   TEXT,   -- full service account JSON, encrypted

    stripe_publishable_key       TEXT,   -- pk_live_... (not secret)
    stripe_secret_key_enc        TEXT,   -- sk_live_..., encrypted
    stripe_webhook_secret_enc    TEXT,   -- whsec_..., encrypted

    paddle_vendor_id             TEXT,
    paddle_api_key_enc           TEXT,   -- encrypted
    paddle_webhook_secret_enc    TEXT,   -- encrypted

    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT app_credentials_unique_provider UNIQUE (app_id, provider)
);

CREATE INDEX idx_app_credentials_app_id ON app_credentials(app_id);

COMMENT ON TABLE app_credentials IS
'Encrypted store credentials per app per provider.
Sensitive fields (*_enc) are AES-256 encrypted by the application before INSERT/UPDATE
and decrypted after SELECT. Never logged, never returned in API responses directly.';

-- ── 3. Seed default settings for existing apps ─────────────────────────────────

UPDATE apps
SET settings = '{
    "grace_period_days": 3,
    "trial_enabled": false,
    "trial_days": 0,
    "default_currency": "USD",
    "store_environment": "production",
    "entitlements": {}
}'::jsonb
WHERE settings = '{}'::jsonb;
