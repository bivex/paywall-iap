# Infrastructure Layer

External concerns: database, cache, external APIs, logging.

## Structure

- `persistence/` - PostgreSQL (sqlc, repositories)
- `cache/` - Redis (rate limiting, caching)
- `external/` - External APIs (Lago, Apple, Google, Stripe)
- `logging/` - Zap logger, Sentry integration
- `metrics/` - Prometheus metrics
- `config/` - Viper configuration

## Dependency Rule

Implements interfaces defined in domain layer.
