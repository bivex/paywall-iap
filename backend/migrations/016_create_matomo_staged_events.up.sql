-- Create matomo_staged_events table for event queue
CREATE TABLE matomo_staged_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type      TEXT NOT NULL CHECK (event_type IN ('event', 'ecommerce', 'custom')),
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    payload         JSONB NOT NULL,

    -- Retry tracking
    retry_count     INT NOT NULL DEFAULT 0,
    max_retries     INT NOT NULL DEFAULT 3,
    next_retry_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Status tracking
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'failed', 'sent')),

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at         TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    error_message   TEXT
);

-- Index for processing pending events
CREATE INDEX idx_matomo_staged_events_pending ON matomo_staged_events(status, next_retry_at)
    WHERE status = 'pending' AND next_retry_at <= now();

-- Index for user's events
CREATE INDEX idx_matomo_staged_events_user ON matomo_staged_events(user_id, created_at DESC);

-- Index for cleanup of old sent events
CREATE INDEX idx_matomo_staged_events_cleanup ON matomo_staged_events(status, sent_at)
    WHERE status = 'sent' AND sent_at < NOW() - INTERVAL '30 days';

-- Index for failed events (for fallback storage)
CREATE INDEX idx_matomo_staged_events_failed ON matomo_staged_events(status, failed_at)
    WHERE status = 'failed' AND retry_count >= max_retries;

-- Comment for documentation
COMMENT ON TABLE matomo_staged_events IS 'Staged Matomo events for async delivery with retry support';
COMMENT ON COLUMN matomo_staged_events.event_type IS 'Type of event: event (standard), ecommerce (purchase), or custom';
COMMENT ON COLUMN matomo_staged_events.payload IS 'Event data as JSONB - category, action, revenue, items, etc.';
COMMENT ON COLUMN matomo_staged_events.next_retry_at IS 'When the event should be retried (exponential backoff)';
COMMENT ON COLUMN matomo_staged_events.status IS 'pending, processing, failed, or sent';

-- Function to calculate next retry time with exponential backoff
CREATE OR REPLACE FUNCTION calculate_next_retry(retry_count INT, max_retries INT) RETURNS TIMESTAMPTZ AS $$
DECLARE
    base_interval INTERVAL := INTERVAL '1 minute';
    max_interval INTERVAL := INTERVAL '1 hour';
    backoff_multiplier INT := 2;
    calculated_interval INTERVAL;
BEGIN
    IF retry_count >= max_retries THEN
        RETURN NULL; -- No more retries
    END IF;

    -- Exponential backoff: 1min, 2min, 4min, 8min, ..., max 1 hour
    calculated_interval := LEAST(
        base_interval * (backoff_multiplier ^ retry_count),
        max_interval
    );

    RETURN now() + calculated_interval;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Trigger to auto-calculate next_retry_at on insert
CREATE OR REPLACE FUNCTION set_matomo_event_initial_retry()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.next_retry_at IS NULL OR NEW.next_retry_at = '1970-01-01 00:00:00+00'::timestamptz THEN
        NEW.next_retry_at = now(); -- Ready for immediate processing
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_matomo_staged_events_insert
    BEFORE INSERT ON matomo_staged_events
    FOR EACH ROW
    EXECUTE FUNCTION set_matomo_event_initial_retry();

-- Function to update event status after successful send
CREATE OR REPLACE FUNCTION mark_matomo_event_sent(event_id UUID) RETURNS VOID AS $$
BEGIN
    UPDATE matomo_staged_events
    SET status = 'sent',
        sent_at = now(),
        next_retry_at = NULL
    WHERE id = event_id;
END;
$$ LANGUAGE plpgsql;

-- Function to mark event as failed and increment retry count
CREATE OR REPLACE FUNCTION mark_matomo_event_failed(event_id UUID, error_msg TEXT) RETURNS VOID AS $$
DECLARE
    current_retry_count INT;
    current_max_retries INT;
BEGIN
    SELECT retry_count, max_retries INTO current_retry_count, current_max_retries
    FROM matomo_staged_events
    WHERE id = event_id;

    IF current_retry_count >= current_max_retries THEN
        -- Permanent failure
        UPDATE matomo_staged_events
        SET status = 'failed',
            failed_at = now(),
            error_message = error_msg
        WHERE id = event_id;
    ELSE
        -- Retry with exponential backoff
        UPDATE matomo_staged_events
        SET retry_count = retry_count + 1,
            status = 'pending',
            next_retry_at = calculate_next_retry(retry_count + 1, max_retries),
            error_message = error_msg
        WHERE id = event_id;
    END IF;
END;
$$ LANGUAGE plpgsql;
