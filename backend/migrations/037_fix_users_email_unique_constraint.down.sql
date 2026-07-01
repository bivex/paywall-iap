-- Restore original indexes (without empty string exclusion)
DROP INDEX IF EXISTS idx_users_app_email;
CREATE UNIQUE INDEX idx_users_app_email
    ON users (app_id, email)
    WHERE email IS NOT NULL;

DROP INDEX IF EXISTS users_email_unique;
CREATE UNIQUE INDEX users_email_unique
    ON users (email)
    WHERE email IS NOT NULL AND email <> '';
