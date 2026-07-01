-- Fix 1: drop global email unique index (replaced by idx_users_app_email which scopes per app_id)
DROP INDEX IF EXISTS users_email_unique;

-- Fix 2: drop single-active-subscription constraint that ignores app_id
DROP INDEX IF EXISTS idx_subscriptions_one_active;

-- Recreate subscription uniqueness scoped to app_id
CREATE UNIQUE INDEX idx_subscriptions_one_active
    ON subscriptions (app_id, user_id)
    WHERE status = 'active' AND deleted_at IS NULL;
