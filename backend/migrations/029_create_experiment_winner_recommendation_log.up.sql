CREATE TABLE experiment_winner_recommendation_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    source TEXT NOT NULL,
    recommended BOOLEAN NOT NULL DEFAULT FALSE,
    reason TEXT NOT NULL,
    winning_arm_id UUID REFERENCES ab_test_arms(id) ON DELETE SET NULL,
    confidence_percent DOUBLE PRECISION,
    confidence_threshold_percent DOUBLE PRECISION NOT NULL,
    observed_samples INTEGER NOT NULL,
    min_sample_size INTEGER NOT NULL,
    details JSONB,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_experiment_winner_recommendation_log_experiment
    ON experiment_winner_recommendation_log(experiment_id, occurred_at DESC);

CREATE INDEX idx_experiment_winner_recommendation_log_source
    ON experiment_winner_recommendation_log(source, occurred_at DESC);

COMMENT ON TABLE experiment_winner_recommendation_log IS 'Append-only history of evaluated winner recommendations for bandit experiments';