-- ============================================================
-- Table: dunning
-- ============================================================

CREATE TABLE dunning (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    status          TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'recovered', 'failed')),
    attempt_count   INTEGER NOT NULL DEFAULT 0,
    max_attempts    INTEGER NOT NULL DEFAULT 5,
    next_attempt_at TIMESTAMPTZ NOT NULL,
    last_attempt_at TIMESTAMPTZ,
    recovered_at    TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE dunning IS 'Subscription dunning process for failed renewals';

CREATE INDEX idx_dunning_subscription
    ON dunning(subscription_id)
    WHERE status IN ('pending', 'in_progress');

CREATE INDEX idx_dunning_pending
    ON dunning(next_attempt_at)
    WHERE status IN ('pending', 'in_progress');
