-- Create ab_test_arms table for experiment variants
CREATE TABLE ab_test_arms (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id   UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    is_control      BOOLEAN NOT NULL DEFAULT false,
    traffic_weight  NUMERIC(3,2) NOT NULL DEFAULT 1.0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index for arm lookup by experiment
CREATE INDEX idx_ab_test_arms_experiment ON ab_test_arms(experiment_id);

-- Index for control arm lookup
CREATE INDEX idx_ab_test_arms_control ON ab_test_arms(experiment_id, is_control) WHERE is_control = true;

-- Ensure traffic_weight is non-negative
CHECK (traffic_weight >= 0)

-- Comment for documentation
COMMENT ON TABLE ab_test_arms IS 'Arms (variants) for A/B test experiments';
COMMENT ON COLUMN ab_test_arms.is_control IS 'True if this is the control/baseline variant';
COMMENT ON COLUMN ab_test_arms.traffic_weight IS 'Initial traffic weight for bandit algorithms';
