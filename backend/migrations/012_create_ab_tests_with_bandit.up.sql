-- Create ab_tests table with bandit support
CREATE TABLE ab_tests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                TEXT NOT NULL,
    description         TEXT,
    status              TEXT NOT NULL CHECK (status IN ('draft', 'running', 'paused', 'completed')) DEFAULT 'draft',
    start_at            TIMESTAMPTZ,
    end_at              TIMESTAMPTZ,

    -- Bandit-specific columns
    algorithm_type      TEXT CHECK (algorithm_type IN ('thompson_sampling', 'ucb', 'epsilon_greedy')),
    is_bandit           BOOLEAN NOT NULL DEFAULT false,
    min_sample_size     INT DEFAULT 100,
    confidence_threshold NUMERIC(3,2) DEFAULT 0.95,
    winner_confidence   NUMERIC(3,2),

    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index for active experiments
CREATE INDEX idx_ab_tests_active ON ab_tests(status) WHERE status = 'running';

-- Comment for documentation
COMMENT ON TABLE ab_tests IS 'A/B test experiments with multi-armed bandit support';
COMMENT ON COLUMN ab_tests.algorithm_type IS 'Bandit algorithm: thompson_sampling, ucb, or epsilon_greedy';
COMMENT ON COLUMN ab_tests.is_bandit IS 'True if this experiment uses bandit optimization (vs fixed allocation)';
COMMENT ON COLUMN ab_tests.min_sample_size IS 'Minimum samples before bandit can declare winner';
COMMENT ON COLUMN ab_tests.confidence_threshold IS 'Statistical confidence threshold (0.95 = 95%)';
COMMENT ON COLUMN ab_tests.winner_confidence IS 'Current confidence level of the winning arm';
