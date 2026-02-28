-- Drop ab_test_arm_stats table
DROP TRIGGER IF EXISTS trg_ab_test_arm_stats_updated_at ON ab_test_arm_stats;
DROP FUNCTION IF EXISTS update_ab_test_arm_stats_updated_at();
DROP INDEX IF EXISTS idx_ab_test_arm_stats_arm;
DROP TABLE IF EXISTS ab_test_arm_stats;
