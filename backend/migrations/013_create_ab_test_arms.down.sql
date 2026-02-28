-- Drop ab_test_arms table
DROP INDEX IF EXISTS idx_ab_test_arms_control;
DROP INDEX IF EXISTS idx_ab_test_arms_experiment;
DROP TABLE IF EXISTS ab_test_arms;
