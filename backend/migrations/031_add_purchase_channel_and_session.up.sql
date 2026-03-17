ALTER TABLE users
  ADD COLUMN purchase_channel TEXT CHECK (purchase_channel IN ('iap', 'stripe', 'web')),
  ADD COLUMN session_count    INT NOT NULL DEFAULT 0,
  ADD COLUMN has_viewed_ads   BOOLEAN NOT NULL DEFAULT false;
