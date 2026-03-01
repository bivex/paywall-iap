-- Allow 'web' platform for admin/backoffice users
ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_platform_check,
  ADD CONSTRAINT users_platform_check
    CHECK (platform IN ('ios', 'android', 'web'));
