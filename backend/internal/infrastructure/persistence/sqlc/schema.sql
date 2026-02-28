-- Canonical schema for sqlc code generation
-- This file is the source of truth for sqlc
-- Keep in sync with migrations

CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_user_id    TEXT UNIQUE NOT NULL,
    device_id           TEXT,
    platform            TEXT NOT NULL CHECK (platform IN ('ios', 'android')),
    app_version         TEXT NOT NULL,
    email               TEXT UNIQUE,
    role                TEXT NOT NULL DEFAULT 'user',
    ltv                 NUMERIC(10,2) DEFAULT 0,
    ltv_updated_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);

CREATE TABLE subscriptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    status          TEXT NOT NULL CHECK (status IN ('active', 'expired', 'cancelled', 'grace')),
    source          TEXT NOT NULL CHECK (source IN ('iap', 'stripe', 'paddle')),
    platform        TEXT NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    product_id      TEXT NOT NULL,
    plan_type       TEXT NOT NULL CHECK (plan_type IN ('monthly', 'annual', 'lifetime')),
    expires_at      TIMESTAMPTZ NOT NULL,
    auto_renew      BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id),
    subscription_id     UUID NOT NULL REFERENCES subscriptions(id),
    amount              NUMERIC(10,2) NOT NULL,
    currency            CHAR(3) NOT NULL,
    status              TEXT NOT NULL CHECK (status IN ('success', 'failed', 'refunded')),
    receipt_hash        TEXT,
    provider_tx_id      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE webhook_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider        TEXT NOT NULL CHECK (provider IN ('stripe', 'apple', 'google', 'paddle')),
    event_type      TEXT NOT NULL,
    event_id        TEXT NOT NULL,
    payload         JSONB NOT NULL,
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_webhook_events_unique
    ON webhook_events(provider, event_id);

CREATE TABLE analytics_aggregates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name     TEXT NOT NULL,
    metric_date     DATE NOT NULL,
    metric_value    NUMERIC(20,2) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_analytics_aggregates_unique
    ON analytics_aggregates(metric_name, metric_date);

CREATE TABLE grace_periods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id),
    status          TEXT NOT NULL CHECK (status IN ('active', 'resolved', 'expired')),
    expires_at      TIMESTAMPTZ NOT NULL,
    resolved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_grace_periods_active
    ON grace_periods(user_id, status)
    WHERE status = 'active';
