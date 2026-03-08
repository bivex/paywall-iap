UPDATE ab_tests
SET automation_policy = automation_policy - 'locked_until' - 'locked_by' - 'lock_reason';

ALTER TABLE ab_tests
ALTER COLUMN automation_policy SET DEFAULT '{
  "enabled": false,
  "auto_start": false,
  "auto_complete": false,
  "complete_on_end_time": true,
  "complete_on_sample_size": false,
  "complete_on_confidence": false,
  "manual_override": false
}'::jsonb;