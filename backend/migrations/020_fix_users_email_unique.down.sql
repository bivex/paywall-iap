-- Revert: restore the original blanket UNIQUE constraint on email
DROP INDEX IF EXISTS users_email_unique;

ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
