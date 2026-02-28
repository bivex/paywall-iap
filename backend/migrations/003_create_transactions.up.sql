-- ============================================================
-- Table: transactions
-- ============================================================
-- receipt_data stores only a SHA-256 hash of the raw receipt
-- for deduplication; the full encrypted receipt is stored as
-- a webhook_event payload.

CREATE TABLE transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id),
    subscription_id     UUID NOT NULL REFERENCES subscriptions(id),
    amount              NUMERIC(10,2) NOT NULL,
    currency            CHAR(3) NOT NULL,
    status              TEXT NOT NULL CHECK (status IN ('success', 'failed', 'refunded')),
    receipt_hash        TEXT,
    provider_tx_id      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE transactions IS 'Payment transactions';
COMMENT ON COLUMN transactions.receipt_hash IS 'SHA-256 hash of receipt for deduplication';

-- Transaction history
CREATE INDEX idx_transactions_user
    ON transactions(user_id, created_at DESC);
