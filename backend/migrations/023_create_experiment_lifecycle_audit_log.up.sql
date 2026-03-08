CREATE TABLE experiment_lifecycle_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id   UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    actor_type      TEXT NOT NULL CHECK (actor_type IN ('admin', 'system')),
    actor_id        UUID,
    source          TEXT NOT NULL,
    action          TEXT NOT NULL CHECK (action IN ('status_transition')),
    from_status     TEXT NOT NULL,
    to_status       TEXT NOT NULL,
    idempotency_key TEXT,
    details         JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_experiment_lifecycle_audit_log_idempotency
    ON experiment_lifecycle_audit_log(idempotency_key);

CREATE INDEX idx_experiment_lifecycle_audit_log_experiment
    ON experiment_lifecycle_audit_log(experiment_id, created_at DESC);

COMMENT ON TABLE experiment_lifecycle_audit_log IS 'Lifecycle audit trail for experiment status transitions from admin and system automation paths';