-- Fix: empty string email was treated as non-NULL, causing UNIQUE constraint
-- violations for users registered without an email on the same app.
-- Rebuild both email-related unique indexes to exclude empty strings.

DROP INDEX IF EXISTS idx_users_app_email;
CREATE UNIQUE INDEX idx_users_app_email
    ON users (app_id, email)
    WHERE email IS NOT NULL AND email <> '';

DROP INDEX IF EXISTS users_email_unique;
CREATE UNIQUE INDEX users_email_unique
    ON users (email)
    WHERE email IS NOT NULL AND email <> '';
