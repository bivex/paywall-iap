UPDATE ab_tests
SET automation_policy = COALESCE(automation_policy, '{}'::jsonb) || '{
  "locked_until": null,
  "locked_by": null,
  "lock_reason": null
}'::jsonb;

ALTER TABLE ab_tests
ALTER COLUMN automation_policy SET DEFAULT '{
  "enabled": false,
  "auto_start": false,
  "auto_complete": false,
  "complete_on_end_time": true,
  "complete_on_sample_size": false,
  "complete_on_confidence": false,
  "manual_override": false,
  "locked_until": null,
  "locked_by": null,
  "lock_reason": null
}'::jsonb;