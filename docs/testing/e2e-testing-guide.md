# E2E Testing Guide

This document defines how End-to-End tests are conducted for the Paywall IAP application.
Our tests utilize Go's built-in `testcontainers-go` module to spin up fully isolated PostgreSQL and (optionally) Redis processes required to run the real HTTP Server and fully process requests.

## Setup

Ensure you have Docker desktop installed and running as it is required for `testcontainers` integration.

## Writing E2E Tests

1. Place your tests in the `backend/tests/e2e/` folder.
2. Ensure you initialize the suite correctly using `suite := SetupE2ETestSuite(ctx, t)` at the start of any new test struct to gain access to the real-world HTTP database client instances.
3. Tests within E2E should not mock the repository (although our mock repo implements the real DB connector SQL, testing exactly how Go communicates to PostgreSQL).
4. Run E2E tests: `make test-e2e`. The runner ensures a clean environment every run by destroying the testcontainer at the end of the test.
