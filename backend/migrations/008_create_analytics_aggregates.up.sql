-- ============================================================
-- Table: analytics_aggregates
-- ============================================================

CREATE TABLE analytics_aggregates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name     TEXT NOT NULL,
    metric_date     DATE NOT NULL,
    metric_value     NUMERIC(20,2) NOT NULL,
    dimensions      JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE analytics_aggregates IS 'Pre-computed analytics aggregates for reporting';

CREATE UNIQUE INDEX idx_analytics_aggregates_unique
    ON analytics_aggregates(metric_name, metric_date, dimensions);
