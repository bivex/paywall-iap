# Testing Strategy

This document outlines the testing strategy for the Paywall IAP project. Our goal is 80% test coverage across the whole application, utilizing four main types of testing.

## Unit Tests
Isolated Go testing on internal domain validation and command handlers. Mock interfaces isolate logic paths.
- Execution: `make test-unit`
- Target coverage: 80% on core commands.

## Integration Tests
Combines database logic (via PostgreSQL testcontainers setup) with query requests or repository implementations instead of using simple in-memory mocks. Used heavily for handlers.
- Execution: `make test-integration`
- Target: Validate database state operations (`UPDATE` cascades, etc).

## E2E Tests
Models the complete user journey: "Registration" -> "Login" -> "Subscribe" -> "Cancel". The HTTP server is launched during tests, making real API calls over localhost:port against testcontainers.
- Execution: `make test-e2e`
- Focus: Prevent critical application-wide regressions.
    
## Load Tests
Checks the application against thresholds defined in SLA via `k6`. Monitors limits, bottlenecks and failures.
- Execution: `make test-load`, `make test-load-stress`, `make test-load-soak`.
- Focus: Latency and error rates under scale-out stress.

## Maintenance Guidelines
- Ensure that for any new model or table created, the inline test migration script `tests/testutil/migrations.go` is appended.
- Tests failing with "Container Failed" require clearing local Docker volume states `docker system prune --volumes`.
