-- ============================================================
-- Table: users
-- ============================================================
-- Platform identity: Apple originalTransactionId or Google
-- obfuscatedExternalAccountId. This is the canonical persistent
-- identity that survives reinstalls and device resets.
-- device_id is stored for analytics only and MUST NOT be used
-- as a foreign key.

CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_user_id    TEXT UNIQUE NOT NULL,
    device_id           TEXT,
    platform            TEXT NOT NULL CHECK (platform IN ('ios', 'android')),
    app_version         TEXT NOT NULL,
    email               TEXT UNIQUE,
    ltv                 NUMERIC(10,2) DEFAULT 0,
    ltv_updated_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);

COMMENT ON TABLE users IS 'User accounts with platform-based identity';
COMMENT ON COLUMN users.platform_user_id IS 'Canonical platform identity (Apple: originalTransactionId, Google: obfuscatedExternalAccountId)';
COMMENT ON COLUMN users.device_id IS 'Device identifier for analytics only - DO NOT use as foreign key';
COMMENT ON COLUMN users.ltv IS 'Lifetime value - updated by worker job';
