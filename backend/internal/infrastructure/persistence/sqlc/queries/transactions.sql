-- name: CreateTransaction :one
INSERT INTO transactions (app_id, user_id, subscription_id, amount, currency, status, receipt_hash, provider_tx_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetTransactionByID :one
SELECT * FROM transactions
WHERE id = $1
LIMIT 1;

-- name: GetTransactionsByUserID :many
SELECT * FROM transactions
WHERE app_id = $1 AND user_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CheckDuplicateReceipt :one
SELECT id FROM transactions
WHERE receipt_hash = $1
LIMIT 1;

-- name: GetLTVByUserID :one
SELECT COALESCE(SUM(amount), 0) AS ltv
FROM transactions
WHERE app_id = $1 AND user_id = $2 AND status = 'success';

-- name: GetTransactionsBySubscriptionID :many
SELECT * FROM transactions
WHERE subscription_id = $1
ORDER BY created_at DESC;

-- name: GetSegmentedLTVByPlatform :many
SELECT u.platform, COALESCE(SUM(t.amount), 0) AS ltv
FROM transactions t
JOIN users u ON t.user_id = u.id
WHERE t.status = 'success'
  AND ($1::int = 0 OR t.created_at >= now() - ($1::int * interval '1 day'))
GROUP BY u.platform;

-- name: GetDailyRevenue :one
SELECT COALESCE(SUM(amount), 0) AS revenue
FROM transactions
WHERE app_id = $1
  AND status = 'success'
  AND created_at >= $2
  AND created_at < $3;
