# Troubleshooting Guide for Tests

## Tests Failing with PostgreSQL Errors
`failed to start container: context deadline exceeded`

### Reason
Docker might not be running locally or the timeout specified within the wait strategy of testcontainers was reached (often Docker takes too long to pull the image).

### Fix
- Ensure Docker is active.
- Pull the image manually once so its fast the second time: `docker pull postgres:15-alpine`.

## Database Duplicate Keys (Flaky tests)
If the testing environment begins spitting out unique identifier constraint collisions across suites:

### Reason
- The tables were not truncated between test cases. Our integration and unit tests often leave behind rows if not handled in teardown correctly, although TestContainers generally isolate environments gracefully between runs.

### Fix
- Utilize the helper `TruncateAll(ctx, pool)` available in `testutil.go`. Do this within a `defer` at the top of an E2E test.
- Check table sequences logic if utilizing sequences `ALTER SEQUENCE ... RESTART WITH 1`.

## Load Test Failures (k6)
If you get `rate<0.05` thresholds failing out on HTTP endpoints.

### Reason
- Base application bottlenecks, frequently Redis backends throttling connections during JWT verifications. (Ensure MaxIdles connections are configured correctly to Redis). The Docker compose defaults don't perform well under k6 scale tests without increasing connection pool sizes.

### Fix
- Update `backend/cmd/api/main.go` and inject higher Go database connection `SetMaxOpenConns()` and `SetMaxIdleConns()` boundaries.
