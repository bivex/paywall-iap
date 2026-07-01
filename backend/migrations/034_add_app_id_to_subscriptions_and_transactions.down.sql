-- Revert 034_add_app_id_to_subscriptions_and_transactions

DROP INDEX IF EXISTS idx_transactions_app;
DROP INDEX IF EXISTS idx_subscriptions_app;

ALTER TABLE transactions  DROP CONSTRAINT IF EXISTS fk_transactions_app;
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS fk_subscriptions_app;

ALTER TABLE transactions  DROP COLUMN IF EXISTS app_id;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS app_id;
