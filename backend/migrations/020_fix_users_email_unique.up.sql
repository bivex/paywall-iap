-- Replace the blanket UNIQUE constraint on email with a partial unique index
-- so that empty-string and NULL emails don't collide with each other,
-- while real email addresses remain globally unique.

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique
    ON users (email)
    WHERE email IS NOT NULL AND email <> '';
