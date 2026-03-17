ALTER TABLE users
  DROP COLUMN IF EXISTS purchase_channel,
  DROP COLUMN IF EXISTS session_count,
  DROP COLUMN IF EXISTS has_viewed_ads;
