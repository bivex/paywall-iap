-- Rollback 036: remove app_settings column and app_credentials table

ALTER TABLE apps DROP CONSTRAINT IF EXISTS apps_settings_grace_period_check;
ALTER TABLE apps DROP CONSTRAINT IF EXISTS apps_settings_trial_days_check;
ALTER TABLE apps DROP COLUMN IF EXISTS settings;

DROP TABLE IF EXISTS app_credentials;
