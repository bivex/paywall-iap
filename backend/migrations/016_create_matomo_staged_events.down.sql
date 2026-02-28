-- Drop matomo_staged_events table and functions
DROP FUNCTION IF EXISTS mark_matomo_event_failed(UUID, TEXT);
DROP FUNCTION IF EXISTS mark_matomo_event_sent(UUID);
DROP TRIGGER IF EXISTS trg_matomo_staged_events_insert ON matomo_staged_events;
DROP FUNCTION IF EXISTS set_matomo_event_initial_retry();
DROP FUNCTION IF EXISTS calculate_next_retry(INT, INT);
DROP INDEX IF EXISTS idx_matomo_staged_events_failed;
DROP INDEX IF EXISTS idx_matomo_staged_events_cleanup;
DROP INDEX IF EXISTS idx_matomo_staged_events_user;
DROP INDEX IF EXISTS idx_matomo_staged_events_pending;
DROP TABLE IF EXISTS matomo_staged_events;
