ALTER TABLE ab_test_arms
    ADD COLUMN pricing_tier_id UUID REFERENCES pricing_tiers(id);

CREATE INDEX idx_ab_test_arms_pricing_tier_id
    ON ab_test_arms(pricing_tier_id)
    WHERE pricing_tier_id IS NOT NULL;

COMMENT ON COLUMN ab_test_arms.pricing_tier_id IS 'Optional pricing tier linked to this experiment arm';