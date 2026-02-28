-- Create ab_test_arm_stats table for bandit statistics
CREATE TABLE ab_test_arm_stats (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    arm_id          UUID NOT NULL UNIQUE REFERENCES ab_test_arms(id) ON DELETE CASCADE,

    -- Thompson Sampling parameters (Beta distribution)
    alpha           NUMERIC(10,2) NOT NULL DEFAULT 1.0,
    beta            NUMERIC(10,2) NOT NULL DEFAULT 1.0,

    -- Sample counts
    samples         INT NOT NULL DEFAULT 0,
    conversions     INT NOT NULL DEFAULT 0,

    -- Revenue tracking
    revenue         NUMERIC(15,2) NOT NULL DEFAULT 0.0,

    -- Computed metrics
    avg_reward      NUMERIC(10,4),

    -- Timestamps
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Constraints
    CHECK (alpha > 0),
    CHECK (beta > 0),
    CHECK (samples >= 0),
    CHECK (conversions >= 0),
    CHECK (conversions <= samples)
);

-- Index for stats lookup by arm
CREATE INDEX idx_ab_test_arm_stats_arm ON ab_test_arm_stats(arm_id);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_ab_test_arm_stats_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_ab_test_arm_stats_updated_at
    BEFORE UPDATE ON ab_test_arm_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_ab_test_arm_stats_updated_at();

-- Comment for documentation
COMMENT ON TABLE ab_test_arm_stats IS 'Bandit statistics for experiment arms (Thompson Sampling parameters)';
COMMENT ON COLUMN ab_test_arm_stats.alpha IS 'Beta distribution alpha parameter (successes + prior)';
COMMENT ON COLUMN ab_test_arm_stats.beta IS 'Beta distribution beta parameter (failures + prior)';
COMMENT ON COLUMN ab_test_arm_stats.avg_reward IS 'Average reward per sample (revenue / samples)';
