-- name: CreateSubscription :one
INSERT INTO subscriptions (user_id, status, source, platform, product_id, plan_type, expires_at, auto_renew)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetActiveSubscriptionByUserID :one
SELECT * FROM subscriptions
WHERE user_id = $1 AND status = 'active' AND deleted_at IS NULL
LIMIT 1;

-- name: GetAccessCheck :one
SELECT id, status, expires_at FROM subscriptions
WHERE user_id = $1
  AND status = 'active'
  AND expires_at > now()
  AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateSubscriptionStatus :one
UPDATE subscriptions
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateSubscriptionExpiry :one
UPDATE subscriptions
SET expires_at = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CancelSubscription :one
UPDATE subscriptions
SET status = 'cancelled', auto_renew = false, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetActiveSubscriptionCount :one
SELECT COUNT(*) FROM subscriptions
WHERE status = 'active'
  AND expires_at > now()
  AND deleted_at IS NULL;

-- name: GetSubscriptionsByUserID :many
SELECT * FROM subscriptions
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;
