-- ============================================================
-- Migration: 033_add_multi_tenancy.down.sql
-- Rolls back the multi-tenancy migration completely.
-- WARNING: This destroys all per-app isolation data.
-- ============================================================

BEGIN;

-- Restore bandit_user_context PK
ALTER TABLE bandit_user_context DROP CONSTRAINT bandit_user_context_pkey;
ALTER TABLE bandit_user_context ADD PRIMARY KEY (user_id);

-- Drop performance indexes
DROP INDEX IF EXISTS idx_users_app_id;
DROP INDEX IF EXISTS idx_pricing_tiers_app_id;
DROP INDEX IF EXISTS idx_ab_tests_app_id;
DROP INDEX IF EXISTS idx_webhook_events_app_id;

-- Restore original unique indexes
DROP INDEX IF EXISTS idx_users_app_platform_user_id;
CREATE UNIQUE INDEX users_platform_user_id_key ON users(platform_user_id);

DROP INDEX IF EXISTS idx_users_app_email;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);

DROP INDEX IF EXISTS idx_pricing_tiers_app_name;
ALTER TABLE pricing_tiers ADD CONSTRAINT pricing_tiers_name_key UNIQUE (name);

DROP INDEX IF EXISTS idx_analytics_aggregates_app_unique;
CREATE UNIQUE INDEX idx_analytics_aggregates_unique
    ON analytics_aggregates(metric_name, metric_date, dimensions);

DROP INDEX IF EXISTS idx_webhook_events_app_unique;
CREATE UNIQUE INDEX idx_webhook_events_unique
    ON webhook_events(provider, event_id);

-- Remove FK constraints
ALTER TABLE users                 DROP CONSTRAINT IF EXISTS fk_users_app;
ALTER TABLE pricing_tiers         DROP CONSTRAINT IF EXISTS fk_pricing_tiers_app;
ALTER TABLE ab_tests              DROP CONSTRAINT IF EXISTS fk_ab_tests_app;
ALTER TABLE webhook_events        DROP CONSTRAINT IF EXISTS fk_webhook_events_app;
ALTER TABLE analytics_aggregates  DROP CONSTRAINT IF EXISTS fk_analytics_app;
ALTER TABLE matomo_staged_events   DROP CONSTRAINT IF EXISTS fk_matomo_app;
ALTER TABLE bandit_user_context    DROP CONSTRAINT IF EXISTS fk_bandit_ctx_app;

-- Remove app_id columns
ALTER TABLE users                 DROP COLUMN app_id;
ALTER TABLE pricing_tiers         DROP COLUMN app_id;
ALTER TABLE ab_tests              DROP COLUMN app_id;
ALTER TABLE webhook_events        DROP COLUMN app_id;
ALTER TABLE analytics_aggregates  DROP COLUMN app_id;
ALTER TABLE matomo_staged_events   DROP COLUMN app_id;
ALTER TABLE bandit_user_context    DROP COLUMN app_id;

-- Drop apps table
DROP TABLE IF EXISTS apps;

COMMIT;
