-- name: GetAdminCredential :one
SELECT user_id, password_hash, created_at, updated_at
FROM admin_credentials
WHERE user_id = $1;

-- name: UpsertAdminCredential :one
INSERT INTO admin_credentials (user_id, password_hash, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (user_id) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    updated_at = now()
RETURNING user_id, password_hash, created_at, updated_at;
