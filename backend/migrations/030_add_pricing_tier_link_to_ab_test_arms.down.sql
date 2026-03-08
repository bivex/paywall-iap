DROP INDEX IF EXISTS idx_ab_test_arms_pricing_tier_id;

ALTER TABLE ab_test_arms
    DROP COLUMN IF EXISTS pricing_tier_id;