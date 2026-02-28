-- name: GetExpiredGracePeriods :many
SELECT * FROM grace_periods
WHERE status = 'active' AND expires_at < now()
ORDER BY created_at ASC;

-- name: UpdateGracePeriodStatus :exec
UPDATE grace_periods
SET status = $2, updated_at = now()
WHERE id = $1;
