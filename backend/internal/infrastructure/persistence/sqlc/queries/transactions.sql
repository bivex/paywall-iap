-- name: CreateTransaction :one
INSERT INTO transactions (user_id, subscription_id, amount, currency, status, receipt_hash, provider_tx_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetTransactionByID :one
SELECT * FROM transactions
WHERE id = $1
LIMIT 1;

-- name: GetTransactionsByUserID :many
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CheckDuplicateReceipt :one
SELECT id FROM transactions
WHERE receipt_hash = $1
LIMIT 1;
