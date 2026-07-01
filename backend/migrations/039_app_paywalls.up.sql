-- Migration 039: app_paywalls — per-app paywall configurations

CREATE TABLE IF NOT EXISTS app_paywalls (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id      UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    definition  JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS app_paywalls_app_id_idx ON app_paywalls(app_id);
CREATE INDEX IF NOT EXISTS app_paywalls_app_id_active_idx ON app_paywalls(app_id, is_active);

-- Only one active paywall per app
CREATE UNIQUE INDEX IF NOT EXISTS app_paywalls_one_active_per_app
    ON app_paywalls(app_id) WHERE is_active = true;

COMMENT ON TABLE app_paywalls IS 'Per-app paywall configurations created in Paywall Creator';
COMMENT ON COLUMN app_paywalls.definition IS 'PaywallDefinition JSON: template, tiers, layout, texts, etc.';
