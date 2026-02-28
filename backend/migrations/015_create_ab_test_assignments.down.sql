-- Drop ab_test_assignments table
DROP FUNCTION IF EXISTS is_assignment_active(TIMESTAMPTZ);
DROP INDEX IF EXISTS idx_ab_test_assignments_user_history;
DROP INDEX IF EXISTS idx_ab_test_assignments_expires;
DROP INDEX IF EXISTS idx_ab_test_assignments_active;
DROP TABLE IF EXISTS ab_test_assignments;
