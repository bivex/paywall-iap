-- ============================================================
-- Table: subscriptions
-- ============================================================

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

COMMENT ON TABLE subscriptions IS 'User subscriptions';
COMMENT ON COLUMN subscriptions.source IS 'Purchase source: iap (Apple/Google), stripe, or paddle';

-- Enforce single active subscription per user
CREATE UNIQUE INDEX idx_subscriptions_one_active
    ON subscriptions(user_id)
    WHERE status = 'active' AND deleted_at IS NULL;

-- Hot path: access check (called on every content open)
CREATE INDEX idx_subscriptions_access
    ON subscriptions(user_id, status, expires_at)
    WHERE deleted_at IS NULL;
