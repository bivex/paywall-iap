-- =====================================================
-- Advanced Thompson Sampling Extensions - Migration 017
-- =====================================================

-- =====================================================
-- Section 1: Currency Conversion
-- =====================================================

-- Currency rates table
CREATE TABLE currency_rates (
    id SERIAL PRIMARY KEY,
    base_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    target_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(18,6) NOT NULL,
    source VARCHAR(50),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(base_currency, target_currency)
);

-- Add currency tracking to arm stats
ALTER TABLE ab_test_arm_stats
ADD COLUMN revenue_usd DECIMAL(18,2) DEFAULT 0,
ADD COLUMN original_currency VARCHAR(3),
ADD COLUMN original_revenue DECIMAL(18,2);

-- Index for currency rate lookups
CREATE INDEX idx_currency_rates_target ON currency_rates(target_currency);
CREATE INDEX idx_currency_rates_updated ON currency_rates(updated_at);

-- =====================================================
-- Section 2: Add window configuration to ab_tests
-- =====================================================

ALTER TABLE ab_tests
ADD COLUMN window_type VARCHAR(10) CHECK (window_type IN ('events', 'time', 'none')),
ADD COLUMN window_size INT DEFAULT 1000,
ADD COLUMN window_min_samples INT DEFAULT 100,
ADD COLUMN objective_type VARCHAR(20) NOT NULL DEFAULT 'conversion' CHECK (objective_type IN ('conversion', 'ltv', 'revenue', 'hybrid')),
ADD COLUMN objective_weights JSONB,
ADD COLUMN price_normalization BOOLEAN DEFAULT FALSE,
ADD COLUMN enable_contextual BOOLEAN DEFAULT FALSE,
ADD COLUMN enable_delayed BOOLEAN DEFAULT FALSE,
ADD COLUMN enable_currency BOOLEAN DEFAULT FALSE,
ADD COLUMN exploration_alpha DECIMAL(4,2) DEFAULT 0.30;

-- =====================================================
-- Section 3: Contextual Bandit (LinUCB)
-- =====================================================

-- User context cache
CREATE TABLE bandit_user_context (
    user_id UUID PRIMARY KEY,
    country VARCHAR(3),
    device VARCHAR(20),
    app_version VARCHAR(20),
    days_since_install INT,
    total_spent DECIMAL(12,2),
    last_purchase_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- LinUCB model parameters
CREATE TABLE bandit_arm_context_model (
    arm_id UUID PRIMARY KEY REFERENCES ab_test_arms(id) ON DELETE CASCADE,
    dimension INT NOT NULL DEFAULT 20,
    matrix_a JSONB NOT NULL DEFAULT '[]',
    vector_b JSONB NOT NULL DEFAULT '[]',
    theta JSONB NOT NULL DEFAULT '[]',
    samples_count BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for context lookups
CREATE INDEX idx_bandit_user_context_updated ON bandit_user_context(updated_at);
CREATE INDEX idx_bandit_arm_context_model_updated ON bandit_arm_context_model(updated_at);

-- =====================================================
-- Section 4: Sliding Window
-- =====================================================

-- Event tracking for sliding window
CREATE TABLE bandit_window_events (
    id BIGSERIAL PRIMARY KEY,
    experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    event_type VARCHAR(10) NOT NULL CHECK (event_type IN ('impression', 'conversion', 'no_conversion')),
    reward_value DECIMAL(12,2),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for window queries
CREATE INDEX idx_bandit_window_events_experiment_arm ON bandit_window_events(experiment_id, arm_id);
CREATE INDEX idx_bandit_window_events_timestamp ON bandit_window_events(timestamp DESC);

-- =====================================================
-- Section 5: Delayed Feedback
-- =====================================================

-- Pending reward tracking
CREATE TABLE bandit_pending_rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    converted BOOLEAN DEFAULT FALSE,
    conversion_value DECIMAL(12,2),
    conversion_currency VARCHAR(3),
    converted_at TIMESTAMPTZ,
    processed_at TIMESTAMPTZ
);

-- Conversion links (pending reward -> transaction)
CREATE TABLE bandit_conversion_links (
    pending_id UUID NOT NULL REFERENCES bandit_pending_rewards(id) ON DELETE CASCADE,
    transaction_id UUID NOT NULL,
    linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pending_id, transaction_id)
);

-- Indexes for pending reward processing
CREATE INDEX idx_bandit_pending_rewards_expires ON bandit_pending_rewards(expires_at) WHERE converted = FALSE;
CREATE INDEX idx_bandit_pending_rewards_user ON bandit_pending_rewards(user_id, experiment_id);
CREATE INDEX idx_bandit_conversion_links_transaction ON bandit_conversion_links(transaction_id);

-- =====================================================
-- Section 6: Multi-Objective Hybrid System
-- =====================================================

-- Per-objective statistics
CREATE TABLE bandit_arm_objective_stats (
    id BIGSERIAL PRIMARY KEY,
    arm_id UUID NOT NULL REFERENCES ab_test_arms(id) ON DELETE CASCADE,
    objective_type VARCHAR(20) NOT NULL CHECK (objective_type IN ('conversion', 'ltv', 'revenue')),
    alpha DECIMAL(10,2) DEFAULT 1.0,
    beta DECIMAL(10,2) DEFAULT 1.0,
    samples BIGINT DEFAULT 0,
    conversions BIGINT DEFAULT 0,
    total_revenue DECIMAL(18,2) DEFAULT 0,
    avg_ltv DECIMAL(12,2),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(arm_id, objective_type)
);

-- Index for objective stats lookups
CREATE INDEX idx_bandit_arm_objective_stats_arm ON bandit_arm_objective_stats(arm_id);
CREATE INDEX idx_bandit_arm_objective_stats_type ON bandit_arm_objective_stats(objective_type);

-- =====================================================
-- Section 7: Helper Functions
-- =====================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_bandit_timestamps()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply timestamp trigger to all relevant tables
CREATE TRIGGER trg_bandit_user_context_updated_at
    BEFORE UPDATE ON bandit_user_context
    FOR EACH ROW
    EXECUTE FUNCTION update_bandit_timestamps();

CREATE TRIGGER trg_bandit_arm_context_model_updated_at
    BEFORE UPDATE ON bandit_arm_context_model
    FOR EACH ROW
    EXECUTE FUNCTION update_bandit_timestamps();

CREATE TRIGGER trg_bandit_arm_objective_stats_updated_at
    BEFORE UPDATE ON bandit_arm_objective_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_bandit_timestamps();

-- =====================================================
-- Section 8: Comments for Documentation
-- =====================================================

COMMENT ON TABLE currency_rates IS 'Exchange rates for currency conversion (USD base)';
COMMENT ON TABLE bandit_user_context IS 'Cached user attributes for contextual bandit features';
COMMENT ON TABLE bandit_arm_context_model IS 'LinUCB model parameters per arm (A matrix, b vector, theta)';
COMMENT ON TABLE bandit_window_events IS 'Event log for sliding window calculations';
COMMENT ON TABLE bandit_pending_rewards IS 'Pending conversions for delayed feedback handling';
COMMENT ON TABLE bandit_conversion_links IS 'Links pending rewards to actual transactions';
COMMENT ON TABLE bandit_arm_objective_stats IS 'Per-objective statistics for multi-objective optimization';

COMMENT ON COLUMN ab_tests.window_type IS 'Type of windowing: events, time, or none';
COMMENT ON COLUMN ab_tests.objective_type IS 'Optimization objective: conversion, ltv, revenue, or hybrid';
COMMENT ON COLUMN ab_tests.objective_weights IS 'JSON weights for hybrid objectives';
COMMENT ON COLUMN ab_tests.enable_contextual IS 'Enable LinUCB contextual bandit';
COMMENT ON COLUMN ab_tests.enable_delayed IS 'Enable delayed feedback handling';
COMMENT ON COLUMN ab_tests.enable_currency IS 'Enable currency conversion to USD';
COMMENT ON COLUMN ab_tests.exploration_alpha IS 'Exploration parameter for LinUCB (default 0.3)';
