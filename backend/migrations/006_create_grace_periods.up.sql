-- ============================================================
-- Table: grace_periods
-- ============================================================

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

COMMENT ON TABLE grace_periods IS 'Subscription grace periods for failed renewals';

CREATE INDEX idx_grace_periods_user
    ON grace_periods(user_id, status)
    WHERE status = 'active';
