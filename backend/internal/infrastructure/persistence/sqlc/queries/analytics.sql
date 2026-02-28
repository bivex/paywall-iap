-- name: UpsertAnalyticsAggregate :exec
INSERT INTO analytics_aggregates (metric_name, metric_date, metric_value)
VALUES ($1, $2, $3)
ON CONFLICT (metric_name, metric_date) DO UPDATE
    SET metric_value = EXCLUDED.metric_value, updated_at = now();
