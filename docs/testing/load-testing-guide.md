# Load Testing Guide

We use [k6](https://k6.io) for load testing the Paywall IAP project to ensure that the API can handle high concurrency gracefully.

## Prerequisites
- Install [k6](https://k6.io/docs/get-started/installation/) locally.
- A running local dev environment (e.g., via `make docker-up`).

## Test Types

1. **Basic Load Test (`make test-load`)**
   - Assesses normal load scenarios (up to 20 users).
2. **Stress Test (`make test-load-stress`)**
   - Ramps up to 200 virtual users to find the breaking points and latency degradation of the API.
3. **Soak Test (`make test-load-soak`)**
   - Runs a steady moderate load (20 users) for 1 hour to identify memory leaks and connection unclosed issues.

## Viewing Results in Grafana
1. A template Grafana dashboard is provided at `backend/tests/load/dashboard.json`.
2. When running `docker-compose.dev.yml` with InfluxDB/Prometheus mapping, k6 can output metrics using:
   `k6 run --out influxdb=http://localhost:8086/k6 tests/load/basic_load_test.js`
3. Import the supplied dashboard JSON into Grafana to see real-time performance.
