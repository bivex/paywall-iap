# Paywall System

In-App Purchase system for iOS and Android with Go backend, Next.js admin dashboard, PostgreSQL, and Redis.

## 🚀 Quick Start

```bash
# Start full stack (API + Worker + DB + Redis + Migrator)
docker compose -f infra/docker-compose/docker-compose.latency-optimized.yml up -d --build

# Start frontend (dev mode — hot reload)
cd frontend && docker compose -f docker-compose.dev.yml up -d --build

# Or frontend production build
cd frontend && docker compose up -d --build
```

## 🔐 Admin Panel

**URL:** http://localhost:3000

| Field | Value |
|-------|-------|
| Email | `admin@paywall.local` |
| Password | `admin12345` |
| Role | `superadmin` |

> ⚠️ Change the password before deploying to production.

### Seed first admin (new DB)

```bash
# Via script (works with Docker DB)
DB_CONTAINER=docker-compose-db-1 ./scripts/seed_admin.sh admin@paywall.local admin12345

# Or via Go command
cd backend && go run ./cmd/seed --email=admin@paywall.local --password=admin12345

# Or via Makefile
cd backend && make seed-admin EMAIL=admin@paywall.local PASSWORD=admin12345
```

### Seed full cold-start test data

```bash
# Docker DB/container mode
DB_CONTAINER=docker-compose-db-1 ./scripts/seed_all_test_data.sh

# Direct Postgres mode
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/iap_db ./scripts/seed_all_test_data.sh
```

This seeds:

- admin credentials
- pricing tiers
- realistic dashboard revenue fixtures
- deterministic experiment/bandit fixtures for A/B Tests, Studio, Delayed Feedback, Sliding Window, and Multi-Objective pages

### Auth flow

```
Browser → POST /auth/v1/login (Next.js server action)
        → POST /v1/admin/auth/login (Go API, bcrypt verify)
        → JWT access token (15 min) + refresh token (30 days)
        → httpOnly cookies set
        → redirect /dashboard/default

/dashboard/* without cookie → proxy.ts → redirect /auth/v1/login
```

## 🏗️ Services

| Service | Port | Description |
|---------|------|-------------|
| Frontend | `3000` | Next.js 16 admin dashboard |
| API | `8080` | Go/Gin REST API |
| Worker | `8081` | Background jobs (River queue) |
| PostgreSQL | `5432` | Main database |
| Redis | `6379` | Cache + JWT blocklist |

## 🗄️ Database

Migrations are applied automatically by the `migrator` container on startup.

```bash
# Apply migrations manually (if needed)
for i in backend/migrations/*.up.sql; do
  docker exec docker-compose-db-1 psql -U postgres -d iap_db < "$i"
done
```

## 🐳 Docker Images

| Service | Compressed | Efficiency |
|---------|-----------|------------|
| API | 14.5 MB | 99% |
| Worker | 11.8 MB | 99% |
| Migrator | 6.9 MB | 98% |
| Frontend (prod) | 83.8 MB | 99.98% |
| Frontend (dev) | 460 MB | 100% |

**Build images:**
```bash
docker build -t paywall-iap-api:latest     -f infra/docker/api/Dockerfile .
docker build -t paywall-iap-worker:latest  -f infra/docker/worker/Dockerfile .
docker build -t paywall-iap-migrator:latest -f infra/docker/migrator/Dockerfile .
```

## ⚡ Performance

```bash
# Latency-optimized stack (BBR, TCP tuning, PG async commit)
docker compose -f infra/docker-compose/docker-compose.latency-optimized.yml up -d
```

See [Latency Optimization Guide](docs/operations/latency-optimization.md) for details.

## 🌐 Production Deploy

Set these env vars before deploying:
```bash
HTTPS=true          # enables Secure flag on cookies
BACKEND_URL=http://api:8080   # internal Docker network
NEXT_PUBLIC_API_URL=https://your-domain.com
```

## 📚 Documentation

- [API Specification](docs/api/openapi.yaml)
- [Database Schema](docs/database/schema-erd.md)
- [Deployment](docs/runbooks/deploy-procedure.md)
- [Latency Optimization](docs/operations/latency-optimization.md)
- [Wireframes](docs/Wireframes_Rethink.md)

