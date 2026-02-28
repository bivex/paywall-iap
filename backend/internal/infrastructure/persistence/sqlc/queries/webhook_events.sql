-- name: InsertWebhookEvent :exec
INSERT INTO webhook_events (provider, event_type, event_id, payload)
VALUES ($1, $2, $3, $4)
ON CONFLICT (provider, event_id) DO NOTHING;

-- name: GetUnprocessedWebhookEvents :many
SELECT * FROM webhook_events
WHERE processed_at IS NULL
ORDER BY created_at ASC
LIMIT 100;

-- name: MarkWebhookEventProcessed :exec
UPDATE webhook_events
SET processed_at = now()
WHERE id = $1;

-- name: GetWebhookEventByProviderAndID :one
SELECT * FROM webhook_events
WHERE provider = $1 AND event_id = $2;
