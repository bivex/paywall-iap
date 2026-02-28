-- ============================================================
-- Table: webhook_events
-- ============================================================
-- Inbox pattern for idempotent webhook processing

CREATE TABLE webhook_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider        TEXT NOT NULL CHECK (provider IN ('stripe', 'apple', 'google', 'paddle')),
    event_type      TEXT NOT NULL,
    event_id        TEXT NOT NULL,
    payload         JSONB NOT NULL,
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE webhook_events IS 'Webhook event inbox for idempotent processing';

-- Unique constraint on provider+event_id for idempotency
CREATE UNIQUE INDEX idx_webhook_events_unique
    ON webhook_events(provider, event_id);

-- Index for unprocessed events
CREATE INDEX idx_webhook_events_unprocessed
    ON webhook_events(processed_at)
    WHERE processed_at IS NULL;
