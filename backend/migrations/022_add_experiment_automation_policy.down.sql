ALTER TABLE ab_tests
DROP CONSTRAINT IF EXISTS ab_tests_automation_policy_is_object,
DROP COLUMN IF EXISTS automation_policy;