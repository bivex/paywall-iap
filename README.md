# IAP System

In-App Purchase system for iOS and Android with Go backend and React Native frontend.

## Quick Start

```bash
# Start local development environment
make dev-up

# Run backend tests
cd backend && make test

# Run mobile tests
cd mobile && npm test
```

## Architecture

Clean Architecture with Go backend API and React Native mobile app.

## Docker Images

Optimized multi-stage builds with stripped binaries for minimal image size.

| Service | Image Size | Compressed | Efficiency |
|---------|------------|------------|------------|
| API | 57 MB | 14.5 MB | 99% |
| Worker | 46 MB | 11.8 MB | 99% |
| Migrator | 25 MB | 6.9 MB | 98% |

**Optimization techniques:**
- Multi-stage builds (Alpine base)
- Stripped binaries (`-ldflags="-s -w"` + `strip`)
- Non-root user execution
- Layer caching with go.mod/prerequisites

Build locally:
```bash
docker build -t paywall-iap-api:latest -f infra/docker/api/Dockerfile .
docker build -t paywall-iap-worker:latest -f infra/docker/worker/Dockerfile .
docker build -t paywall-iap-migrator:latest -f infra/docker/migrator/Dockerfile .
```

## Documentation

- [API Specification](docs/api/openapi.yaml)
- [Database Schema](docs/database/schema-erd.md)
- [Deployment](docs/runbooks/deploy-procedure.md)
