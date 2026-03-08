CREATE TABLE automation_job_run_log (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_name                TEXT NOT NULL,
    source                  TEXT NOT NULL,
    idempotency_key         TEXT NOT NULL,
    status                  TEXT NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    payload                 JSONB,
    details                 JSONB,
    window_started_at       TIMESTAMPTZ NOT NULL,
    window_duration_seconds INTEGER NOT NULL CHECK (window_duration_seconds > 0),
    started_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at             TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_automation_job_run_log_idempotency
    ON automation_job_run_log(idempotency_key);

CREATE INDEX idx_automation_job_run_log_job_window
    ON automation_job_run_log(job_name, window_started_at DESC);

COMMENT ON TABLE automation_job_run_log IS 'Persisted execution log and idempotency claim table for scheduled automation and maintenance jobs';