-- name: CreateUser :one
INSERT INTO users (platform_user_id, device_id, platform, app_version, email, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads;


-- name: GetUserByID :one
SELECT id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads FROM users
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByPlatformID :one
SELECT id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads FROM users
WHERE platform_user_id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads FROM users
WHERE email = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateUserLTV :one
UPDATE users
SET ltv = $2, ltv_updated_at = now()
WHERE id = $1
RETURNING id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads;

-- name: SoftDeleteUser :one
UPDATE users
SET deleted_at = now()
WHERE id = $1
RETURNING id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads;

-- name: ListUsers :many
SELECT id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL;

-- name: UpdateUserRole :one
UPDATE users
SET role = $2
WHERE id = $1
RETURNING id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads;

-- name: UpdateUserPurchaseChannel :one
UPDATE users
SET purchase_channel = $2
WHERE id = $1
RETURNING id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads;

-- name: UpdateUserEmail :one
UPDATE users
SET email = $2
WHERE id = $1
RETURNING id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads;

-- name: IncrementUserSessionCount :one
UPDATE users
SET session_count = session_count + 1
WHERE id = $1
RETURNING id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads;

-- name: UpdateUserHasViewedAds :one
UPDATE users
SET has_viewed_ads = $2
WHERE id = $1
RETURNING id, platform_user_id, device_id, platform, app_version, email, role, ltv, ltv_updated_at, created_at, deleted_at, purchase_channel, session_count, has_viewed_ads;
