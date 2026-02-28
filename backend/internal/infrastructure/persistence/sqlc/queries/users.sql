-- name: CreateUser :one
INSERT INTO users (platform_user_id, device_id, platform, app_version, email)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByPlatformID :one
SELECT * FROM users
WHERE platform_user_id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateUserLTV :one
UPDATE users
SET ltv = $2, ltv_updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SoftDeleteUser :one
UPDATE users
SET deleted_at = now()
WHERE id = $1
RETURNING *;
