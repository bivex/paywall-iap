-- ============================================================
-- Table: pricing_tiers
-- ============================================================

CREATE TABLE pricing_tiers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL UNIQUE,
    description     TEXT,
    monthly_price   NUMERIC(10,2),
    annual_price    NUMERIC(10,2),
    currency        CHAR(3) NOT NULL DEFAULT 'USD',
    features        JSONB,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

COMMENT ON TABLE pricing_tiers IS 'Product pricing tiers';
