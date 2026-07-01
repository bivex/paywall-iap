-- Migration: 034_add_app_id_to_subscriptions_and_transactions
-- Purpose:   Complete multi-tenancy (started in 033) for the two app-scoped
--             tables that were missed: subscriptions and transactions.
--             Their sqlc queries already filter by app_id, so the missing
--             columns caused HTTP 500 ("column app_id does not exist") on
--             /v1/admin/analytics/report and /v1/admin/transactions.
-- Pattern:    matches 033_add_multi_tenancy (add nullable -> back-fill -> NOT NULL -> FK -> index).

-- Step 1 — add app_id columns (nullable first for back-fill)
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS app_id UUID;
ALTER TABLE transactions  ADD COLUMN IF NOT EXISTS app_id UUID;

-- Step 2 — back-fill from each row's owning user's app
UPDATE subscriptions s
   SET app_id = u.app_id
  FROM users u
 WHERE s.user_id = u.id
   AND s.app_id IS NULL;

UPDATE transactions t
   SET app_id = u.app_id
  FROM users u
 WHERE t.user_id = u.id
   AND t.app_id IS NULL;

-- Step 3 — NOT NULL + FK + index (matches 033)
ALTER TABLE subscriptions ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE transactions  ALTER COLUMN app_id SET NOT NULL;

ALTER TABLE subscriptions
  ADD CONSTRAINT fk_subscriptions_app FOREIGN KEY (app_id) REFERENCES apps(id);
ALTER TABLE transactions
  ADD CONSTRAINT fk_transactions_app FOREIGN KEY (app_id) REFERENCES apps(id);

CREATE INDEX IF NOT EXISTS idx_subscriptions_app ON subscriptions (app_id);
CREATE INDEX IF NOT EXISTS idx_transactions_app  ON transactions (app_id);
