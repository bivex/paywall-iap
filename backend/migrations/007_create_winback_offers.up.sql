-- ============================================================
-- Table: winback_offers
-- ============================================================

CREATE TABLE winback_offers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    campaign_id     TEXT NOT NULL,
    discount_type   TEXT NOT NULL CHECK (discount_type IN ('percentage', 'fixed')),
    discount_value  NUMERIC(10,2) NOT NULL,
    status          TEXT NOT NULL CHECK (status IN ('offered', 'accepted', 'expired', 'declined')),
    offered_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL,
    accepted_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE winback_offers IS 'Winback discount offers for churned users';

CREATE UNIQUE INDEX idx_winback_offers_user_campaign
    ON winback_offers(user_id, campaign_id)
    WHERE status IN ('offered', 'accepted');
