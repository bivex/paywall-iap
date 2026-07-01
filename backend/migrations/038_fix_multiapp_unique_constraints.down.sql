-- Restore global email unique index
CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique
    ON users (email)
    WHERE email IS NOT NULL AND email <> '';

-- Restore original subscription uniqueness (without app_id scope)
DROP INDEX IF EXISTS idx_subscriptions_one_active;
CREATE UNIQUE INDEX idx_subscriptions_one_active
    ON subscriptions (user_id)
    WHERE status = 'active' AND deleted_at IS NULL;
