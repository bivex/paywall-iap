-- ============================================================
-- Migration: 033_add_multi_tenancy.up.sql
-- Purpose:   Add app_id (multi-tenancy) support for Mothsalt
--            company running 5 apps (3 iOS, 2 Android) on a
--            single paywall-iap instance.
--
-- Strategy:
--   1. Create `apps` table — one row per Mothsalt application.
--   2. Add `app_id UUID NOT NULL` to every table that holds
--      per-app data (users, pricing_tiers, ab_tests and their
--      descendants, analytics, webhooks).
--   3. Back-fill existing rows with a default sentinel app so
--      the migration is safe on a live DB.
--   4. Re-create all unique indexes that must be app-scoped.
--   5. Add FK constraints after back-fill.
--
-- Tables that do NOT need app_id (global / infra):
--   admin_credentials, admin_audit_log, admin_settings,
--   automation_job_run_log, currency_rates,
--   experiment_lifecycle_audit_log,
--   experiment_automation_decision_log (experiment-scoped, inherits via FK)
-- ============================================================

BEGIN;

-- ============================================================
-- Step 1 — apps table
-- ============================================================

CREATE TABLE apps (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL UNIQUE,          -- e.g. 'com.mothsalt.game1'
    display_name TEXT NOT NULL,                 -- e.g. 'Mothsalt Game 1'
    platform     TEXT NOT NULL CHECK (platform IN ('ios', 'android', 'both')),
    bundle_id    TEXT NOT NULL UNIQUE,          -- App Store / Google Play bundle
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE apps IS 'Mothsalt registered applications — one row per app, used for multi-tenancy isolation';
COMMENT ON COLUMN apps.bundle_id IS 'App Store bundle ID or Google Play package name — used as the app identifier in JWT claims';

-- Seed the 5 Mothsalt apps
INSERT INTO apps (name, display_name, platform, bundle_id) VALUES
    ('com.mothsalt.game1',    'Mothsalt Game 1',    'ios',     'com.mothsalt.game1'),
    ('com.mothsalt.game2',    'Mothsalt Game 2',    'ios',     'com.mothsalt.game2'),
    ('com.mothsalt.game3',    'Mothsalt Game 3',    'ios',     'com.mothsalt.game3'),
    ('com.mothsalt.game4',    'Mothsalt Game 4',    'android', 'com.mothsalt.game4'),
    ('com.mothsalt.game5',    'Mothsalt Game 5',    'android', 'com.mothsalt.game5');

-- Sentinel app for existing data back-fill
INSERT INTO apps (id, name, display_name, platform, bundle_id) VALUES
    ('00000000-0000-0000-0000-000000000001',
     'com.mothsalt.legacy', 'Legacy (pre-migration)', 'both', 'com.mothsalt.legacy');

-- ============================================================
-- Step 2 — Add app_id columns (nullable first for back-fill)
-- ============================================================

ALTER TABLE users            ADD COLUMN app_id UUID;
ALTER TABLE pricing_tiers    ADD COLUMN app_id UUID;
ALTER TABLE ab_tests         ADD COLUMN app_id UUID;
ALTER TABLE webhook_events   ADD COLUMN app_id UUID;
ALTER TABLE analytics_aggregates ADD COLUMN app_id UUID;
ALTER TABLE matomo_staged_events ADD COLUMN app_id UUID;
ALTER TABLE bandit_user_context  ADD COLUMN app_id UUID;

-- ============================================================
-- Step 3 — Back-fill existing rows with the legacy sentinel
-- ============================================================

UPDATE users             SET app_id = '00000000-0000-0000-0000-000000000001';
UPDATE pricing_tiers     SET app_id = '00000000-0000-0000-0000-000000000001';
UPDATE ab_tests          SET app_id = '00000000-0000-0000-0000-000000000001';
UPDATE webhook_events    SET app_id = '00000000-0000-0000-0000-000000000001';
UPDATE analytics_aggregates SET app_id = '00000000-0000-0000-0000-000000000001';
UPDATE matomo_staged_events  SET app_id = '00000000-0000-0000-0000-000000000001';
UPDATE bandit_user_context   SET app_id = '00000000-0000-0000-0000-000000000001';

-- ============================================================
-- Step 4 — Set NOT NULL + add FK constraints
-- ============================================================

ALTER TABLE users            ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE pricing_tiers    ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE ab_tests         ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE webhook_events   ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE analytics_aggregates ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE matomo_staged_events  ALTER COLUMN app_id SET NOT NULL;
ALTER TABLE bandit_user_context   ALTER COLUMN app_id SET NOT NULL;

ALTER TABLE users            ADD CONSTRAINT fk_users_app            FOREIGN KEY (app_id) REFERENCES apps(id);
ALTER TABLE pricing_tiers    ADD CONSTRAINT fk_pricing_tiers_app    FOREIGN KEY (app_id) REFERENCES apps(id);
ALTER TABLE ab_tests         ADD CONSTRAINT fk_ab_tests_app         FOREIGN KEY (app_id) REFERENCES apps(id);
ALTER TABLE webhook_events   ADD CONSTRAINT fk_webhook_events_app   FOREIGN KEY (app_id) REFERENCES apps(id);
ALTER TABLE analytics_aggregates ADD CONSTRAINT fk_analytics_app    FOREIGN KEY (app_id) REFERENCES apps(id);
ALTER TABLE matomo_staged_events  ADD CONSTRAINT fk_matomo_app      FOREIGN KEY (app_id) REFERENCES apps(id);
ALTER TABLE bandit_user_context   ADD CONSTRAINT fk_bandit_ctx_app  FOREIGN KEY (app_id) REFERENCES apps(id);

-- ============================================================
-- Step 5 — Re-create unique indexes as (app_id, ...)
-- ============================================================

-- users: platform_user_id was globally unique; now unique per app
ALTER TABLE users DROP CONSTRAINT users_platform_user_id_key;
CREATE UNIQUE INDEX idx_users_app_platform_user_id
    ON users(app_id, platform_user_id);

-- users: email unique per app (same email can exist in game1 and game2)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX idx_users_app_email
    ON users(app_id, email)
    WHERE email IS NOT NULL;

-- pricing_tiers: name unique per app
ALTER TABLE pricing_tiers DROP CONSTRAINT pricing_tiers_name_key;
CREATE UNIQUE INDEX idx_pricing_tiers_app_name
    ON pricing_tiers(app_id, name)
    WHERE deleted_at IS NULL;

-- subscriptions: one active subscription per user (already user-scoped,
-- user is now app-scoped so this constraint is still correct — no change needed)

-- analytics_aggregates: unique metric per app
DROP INDEX IF EXISTS idx_analytics_aggregates_unique;
CREATE UNIQUE INDEX idx_analytics_aggregates_app_unique
    ON analytics_aggregates(app_id, metric_name, metric_date, dimensions);

-- webhook_events: idempotency key is provider+event_id, globally unique
-- (webhooks arrive from Apple/Google per bundle_id, so add app_id)
DROP INDEX idx_webhook_events_unique;
CREATE UNIQUE INDEX idx_webhook_events_app_unique
    ON webhook_events(app_id, provider, event_id);

-- ============================================================
-- Step 6 — Performance indexes for hot paths
-- ============================================================

CREATE INDEX idx_users_app_id          ON users(app_id);
CREATE INDEX idx_pricing_tiers_app_id  ON pricing_tiers(app_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_ab_tests_app_id       ON ab_tests(app_id);
CREATE INDEX idx_webhook_events_app_id ON webhook_events(app_id);

-- bandit_user_context: PK was user_id; now it's (app_id, user_id)
ALTER TABLE bandit_user_context DROP CONSTRAINT bandit_user_context_pkey;
ALTER TABLE bandit_user_context ADD PRIMARY KEY (app_id, user_id);

COMMIT;
