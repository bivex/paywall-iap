ALTER TABLE ab_tests
ADD COLUMN automation_policy JSONB NOT NULL DEFAULT '{
  "enabled": false,
  "auto_start": false,
  "auto_complete": false,
  "complete_on_end_time": true,
  "complete_on_sample_size": false,
  "complete_on_confidence": false,
  "manual_override": false
}'::jsonb,
ADD CONSTRAINT ab_tests_automation_policy_is_object CHECK (jsonb_typeof(automation_policy) = 'object');

COMMENT ON COLUMN ab_tests.automation_policy IS 'Persisted automation rules for scheduler-driven experiment lifecycle management';