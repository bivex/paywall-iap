-- ============================================================
-- Table: admin_audit_log
-- ============================================================

CREATE TABLE admin_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id        UUID NOT NULL,
    action          TEXT NOT NULL,
    target_user_id  UUID,
    target_type     TEXT NOT NULL,
    details         JSONB,
    ip_address      TEXT,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE admin_audit_log IS 'Audit log for admin actions';

CREATE INDEX idx_admin_audit_log_admin
    ON admin_audit_log(admin_id, created_at DESC);
CREATE INDEX idx_admin_audit_log_target_user
    ON admin_audit_log(target_user_id, created_at DESC);
