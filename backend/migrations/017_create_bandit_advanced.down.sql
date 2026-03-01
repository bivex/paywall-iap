-- =====================================================
-- Advanced Thompson Sampling Extensions - Rollback 017
-- =====================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trg_bandit_arm_objective_stats_updated_at ON bandit_arm_objective_stats;
DROP TRIGGER IF EXISTS trg_bandit_arm_context_model_updated_at ON bandit_arm_context_model;
DROP TRIGGER IF EXISTS trg_bandit_user_context_updated_at ON bandit_user_context;

-- Drop function
DROP FUNCTION IF EXISTS update_bandit_timestamps;

-- Drop Section 6: Multi-Objective Hybrid System
DROP TABLE IF EXISTS bandit_arm_objective_stats;

-- Drop Section 5: Delayed Feedback
DROP TABLE IF EXISTS bandit_conversion_links;
DROP TABLE IF EXISTS bandit_pending_rewards;

-- Drop Section 4: Sliding Window
DROP TABLE IF EXISTS bandit_window_events;

-- Drop Section 3: Contextual Bandit (LinUCB)
DROP TABLE IF EXISTS bandit_arm_context_model;
DROP TABLE IF EXISTS bandit_user_context;

-- Drop Section 2: Window configuration from ab_tests
ALTER TABLE ab_tests
DROP COLUMN IF EXISTS exploration_alpha,
DROP COLUMN IF EXISTS enable_currency,
DROP COLUMN IF EXISTS enable_delayed,
DROP COLUMN IF EXISTS enable_contextual,
DROP COLUMN IF EXISTS price_normalization,
DROP COLUMN IF EXISTS objective_weights,
DROP COLUMN IF EXISTS objective_type,
DROP COLUMN IF EXISTS window_min_samples,
DROP COLUMN IF EXISTS window_size,
DROP COLUMN IF EXISTS window_type;

-- Drop Section 1: Currency Conversion
ALTER TABLE ab_test_arm_stats
DROP COLUMN IF EXISTS original_revenue,
DROP COLUMN IF EXISTS original_currency,
DROP COLUMN IF EXISTS revenue_usd;

DROP TABLE IF EXISTS currency_rates;
