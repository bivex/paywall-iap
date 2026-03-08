CREATE TABLE bandit_impression_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    event_type TEXT NOT NULL CHECK (event_type IN ('impression')),
    metadata JSONB,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bandit_impression_events_experiment
    ON bandit_impression_events(experiment_id, occurred_at DESC);

CREATE INDEX idx_bandit_impression_events_user
    ON bandit_impression_events(user_id, occurred_at DESC);

CREATE INDEX idx_bandit_impression_events_arm
    ON bandit_impression_events(arm_id, occurred_at DESC);

COMMENT ON TABLE bandit_impression_events IS 'Append-only impression history for bandit arm exposures';