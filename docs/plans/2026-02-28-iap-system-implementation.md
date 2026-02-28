# IAP System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a standalone In-App Purchase system for iOS and Android with Go backend API, React Native mobile app, PostgreSQL, Redis, and full subscription/revenue management.

**Architecture:** Clean Architecture (Presentation → Application → Domain ← Infrastructure). Mobile app sends events only; Go API makes all decisions. Workers compute analytics on schedule. All external service calls have defined degradation behavior.

**Tech Stack:**
- Backend: Go 1.21+, Gin, go-iap, pgx/v5, sqlc, Redis, Asynq, Zap
- Frontend: React Native 0.73+, react-native-iap, Zustand, TypeScript
- Infrastructure: PostgreSQL 15, Redis 7.2, Docker Compose, Nginx

---

## Phase 1: Project Setup

### Task 1: Create Repository Structure and Initial Files

**Files:**
- Create: `.gitignore`
- Create: `.editorconfig`
- Create: `README.md`
- Create: `.env.example`

**Step 1: Create .gitignore**

```bash
# .gitignore
cat > .gitignore << 'EOF'
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Go
go.work
go.work.sum
*.sum

# Node
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*
.pnp.*
.pnp.js

# Mobile
mobile/ios/Pods/
mobile/ios/build/
mobile/android/build/
mobile/.gradle/
*.ipa
*.apk

# Environment
.env
.env.local
.env.*.local
*.env

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Generated
backend/internal/infrastructure/persistence/sqlc/generated/
mobile/src/generated/

# Secrets
*.key
*.pem
*.crt
secrets/
EOF
```

**Step 2: Create .editorconfig**

```bash
# .editorconfig
cat > .editorconfig << 'EOF'
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.{go,mod,sum}]
indent_style = tab
indent_size = 4

[*.{ts,tsx,js,jsx,json,yaml,yml}]
indent_style = space
indent_size = 2

[*.md]
trim_trailing_whitespace = false
EOF
```

**Step 3: Create README.md**

```bash
# README.md
cat > README.md << 'EOF'
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

## Documentation

- [API Specification](docs/api/openapi.yaml)
- [Database Schema](docs/database/schema-erd.md)
- [Deployment](docs/runbooks/deploy-procedure.md)
EOF
```

**Step 4: Create .env.example**

```bash
# .env.example
cat > .env.example << 'EOF'
# Database
DATABASE_URL=postgresql://appuser:CHANGE_ME@localhost:5432/iap_db
DB_USER=appuser
DB_NAME=iap_db
DB_PASSWORD=CHANGE_ME

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=CHANGE_ME

# Auth
JWT_SECRET=CHANGE_ME_min_32_chars
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h

# External - IAP
APPLE_SHARED_SECRET=CHANGE_ME
GOOGLE_SERVICE_ACCOUNT_JSON=/run/secrets/google-service-account.json

# External - Billing
LAGO_API_KEY=CHANGE_ME
LAGO_WEBHOOK_SECRET=CHANGE_ME

# External - Payments
STRIPE_SECRET_KEY=sk_test_CHANGE_ME
STRIPE_WEBHOOK_SECRET=whsec_CHANGE_ME
PADDLE_API_KEY=CHANGE_ME
PADDLE_WEBHOOK_SECRET=CHANGE_ME

# Backup
BACKUP_ENCRYPTION_KEY=CHANGE_ME_min_32_chars
BACKUP_BUCKET=your-backup-bucket
KMS_KEY_ID=arn:aws:kms:REGION:ACCOUNT:key/KEY_ID

# mTLS
MTLS_CA_CERT=/run/secrets/ca.crt
MTLS_CLIENT_CERT=/run/secrets/client.crt
MTLS_CLIENT_KEY=/run/secrets/client.key

# Sentry
SENTRY_DSN=https://CHANGE_ME@sentry.io/PROJECT_ID
EOF
```

**Step 5: Commit**

```bash
git add .gitignore .editorconfig README.md .env.example
git commit -m "chore: add initial project configuration files"
```

---

### Task 2: Create Backend Directory Structure

**Files:**
- Create: `backend/internal/domain/README.md`
- Create: `backend/internal/application/README.md`
- Create: `backend/internal/infrastructure/README.md`
- Create: `backend/internal/interfaces/README.md`

**Step 1: Create directory structure**

```bash
mkdir -p backend/{cmd/{api,worker,migrator},internal/{domain/{entity,valueobject,repository,service,event,errors},application/{command,query,dto,middleware,validator},infrastructure/{persistence/{sqlc/queries,pool,repository},cache/{redis,rate_limiter,repository},external/{lago,iap,stripe,paddle},logging,metrics,config},interfaces/{http/{handlers,middleware,response},webhook/{handlers,middleware,events}}},tests/{unit,integration,e2e,load,mocks,testdata},pkg/{errors,utils,constants},migrations}
```

**Step 2: Create layer README files**

```bash
# backend/internal/domain/README.md
cat > backend/internal/domain/README.md << 'EOF'
# Domain Layer

Core business logic and enterprise rules.

## Structure

- `entity/` - Business entities (User, Subscription, Transaction, etc.)
- `valueobject/` - Value objects (Money, Email, PlanType, etc.)
- `repository/` - Repository interfaces (implemented by infrastructure)
- `service/` - Domain services (pricing, eligibility, churn risk)
- `event/` - Domain events
- `errors/` - Domain-specific errors

## Dependency Rule

The domain layer has NO dependencies on other layers.
EOF
```

```bash
# backend/internal/application/README.md
cat > backend/internal/application/README.md << 'EOF'
# Application Layer

Use cases, commands, queries, and DTOs.

## Structure

- `command/` - Command handlers (write operations)
- `query/` - Query handlers (read operations)
- `dto/` - Data transfer objects
- `middleware/` - Application middleware
- `validator/` - Request validators

## Dependency Rule

Depends on domain layer only.
EOF
```

```bash
# backend/internal/infrastructure/README.md
cat > backend/internal/infrastructure/README.md << 'EOF'
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
EOF
```

```bash
# backend/internal/interfaces/README.md
cat > backend/internal/interfaces/README.md << 'EOF'
# Interfaces Layer

HTTP handlers, webhooks, middleware.

## Structure

- `http/` - HTTP API (handlers, middleware, response writer)
- `webhook/` - Webhook handlers (Stripe, Apple, Google, Lago)

## Dependency Rule

Depends on application layer. Delegates business logic to use cases.
EOF
```

**Step 3: Verify structure**

```bash
tree -L 4 backend/internal/
```

Expected: Directory tree with all layer directories created

**Step 4: Commit**

```bash
git add backend/internal/
git commit -m "chore: create backend directory structure with clean architecture layers"
```

---

### Task 3: Initialize Go Module and Dependencies

**Files:**
- Create: `backend/go.mod`
- Create: `backend/go.sum` (generated)
- Create: `backend/Makefile`

**Step 1: Initialize go.mod**

```bash
cd backend
go mod init github.com/your-org/iap-system
```

**Step 2: Create go.mod with dependencies**

```bash
# backend/go.mod
cat > go.mod << 'EOF'
module github.com/your-org/iap-system

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-redis/redis_rate/v10 v10.0.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/jackc/pgx/v5 v5.5.0
	github.com/hibiken/asynq v0.24.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/stretchr/testify v1.8.4
	go.uber.org/zap v1.26.0
	github.com/prometheus/client_golang v1.17.0
	github.com/spf13/viper v1.18.0
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/google/uuid v1.5.0
)
EOF
```

**Step 3: Download dependencies**

```bash
go mod download
go mod tidy
```

Expected: Dependencies downloaded, go.sum created

**Step 4: Create Makefile**

```bash
# backend/Makefile
cat > Makefile << 'EOF'
.PHONY: build test test-unit test-integration test-coverage test-load lint fmt migrate sqlc docker-up docker-down

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker
	go build -o bin/migrator ./cmd/migrator

sqlc:
	sqlc generate

test:
	go test ./... -race -count=1

test-unit:
	go test ./internal/domain/... ./internal/application/... -race

test-integration:
	go test ./tests/integration/... -tags=integration -race

test-coverage:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@awk '/^total:/ { if ($$3+0 < 80) { print "Coverage below 80%: " $$3; exit 1 } }' coverage.out

migrate:
	go run ./cmd/migrator migrate up

migrate-down:
	go run ./cmd/migrator migrate down 1

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .

docker-up:
	docker compose -f ../infra/docker-compose/docker-compose.dev.yml up -d

docker-down:
	docker compose -f ../infra/docker-compose/docker-compose.dev.yml down
EOF
```

**Step 5: Verify**

```bash
go mod verify
```

Expected: All modules verified

**Step 6: Commit**

```bash
git add backend/go.mod backend/go.sum backend/Makefile
git commit -m "chore: initialize Go module with core dependencies"
```

---

### Task 4: Create Infrastructure Docker Configuration

**Files:**
- Create: `infra/docker-compose/docker-compose.yml`
- Create: `infra/docker-compose/docker-compose.dev.yml`
- Create: `infra/docker-compose/docker-compose.test.yml`
- Create: `infra/docker-compose/docker-compose.monitoring.yml`

**Step 1: Create production docker-compose.yml**

```bash
# infra/docker-compose/docker-compose.yml
cat > infra/docker-compose/docker-compose.yml << 'EOF'
version: '3.8'

services:
  api:
    image: registry.yourapp.com/iap-api:${IMAGE_TAG:-latest}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 15s
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - LAGO_API_KEY=${LAGO_API_KEY}
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    networks:
      - internal
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'

  worker:
    image: registry.yourapp.com/iap-worker:${IMAGE_TAG:-latest}
    restart: unless-stopped
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    networks:
      - internal
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.5'

  db:
    image: postgres:15-alpine
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 10s
      timeout: 5s
      retries: 5
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_DB=${DB_NAME}
    networks:
      - internal
    deploy:
      resources:
        limits:
          memory: 2G

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - redisdata:/data
    networks:
      - internal
    deploy:
      resources:
        limits:
          memory: 512M

volumes:
  pgdata:
  redisdata:

networks:
  internal:
    internal: true
EOF
```

**Step 2: Create development docker-compose.dev.yml**

```bash
# infra/docker-compose/docker-compose.dev.yml
cat > infra/docker-compose/docker-compose.dev.yml << 'EOF'
version: '3.8'

services:
  db:
    image: postgres:15-alpine
    restart: unless-stopped
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d iap_db"]
      interval: 5s
      timeout: 3s
      retries: 5
    volumes:
      - pgdata_dev:/var/lib/postgresql/data
    environment:
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_USER=postgres
      - POSTGRES_DB=iap_db
    networks:
      - internal

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redisdata_dev:/data
    networks:
      - internal

volumes:
  pgdata_dev:
  redisdata_dev:

networks:
  internal:
EOF
```

**Step 3: Create test docker-compose.test.yml**

```bash
# infra/docker-compose/docker-compose.test.yml
cat > infra/docker-compose/docker-compose.test.yml << 'EOF'
version: '3.8'

services:
  db-test:
    image: postgres:15-alpine
    restart: unless-stopped
    ports:
      - "5433:5432"
    environment:
      - POSTGRES_PASSWORD=test
      - POSTGRES_USER=test
      - POSTGRES_DB=iap_test
    networks:
      - test

networks:
  test:
EOF
```

**Step 4: Create monitoring docker-compose.monitoring.yml**

```bash
# infra/docker-compose/docker-compose.monitoring.yml
cat > infra/docker-compose/docker-compose.monitoring.yml << 'EOF'
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:v2.47.0
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - ./monitoring/prometheus/rules:/etc/prometheus/rules:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - monitoring

  grafana:
    image: grafana/grafana:10.2.0
    restart: unless-stopped
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources:ro
      - grafana_data:/var/lib/grafana
    networks:
      - monitoring

  postgres-exporter:
    image: prometheuscommunity/postgres-exporter:v0.15.0
    restart: unless-stopped
    environment:
      - DATA_SOURCE_NAME=postgresql://postgres:postgres@db:5432/iap_db?sslmode=disable
    networks:
      - monitoring

  redis-exporter:
    image: oliver006/redis_exporter:v1.54.0
    restart: unless-stopped
    environment:
      - REDIS_ADDR=redis://redis:6379
    networks:
      - monitoring

volumes:
  prometheus_data:
  grafana_data:

networks:
  monitoring:
EOF
```

**Step 5: Create .env for local development**

```bash
cp .env.example infra/docker-compose/.env
```

**Step 6: Verify docker-compose files**

```bash
docker compose -f infra/docker-compose/docker-compose.dev.yml config
```

Expected: Valid configuration output

**Step 7: Commit**

```bash
git add infra/docker-compose/
git commit -m "chore: add Docker Compose configurations for dev, test, prod, and monitoring"
```

---

### Task 5: Set Up CI/CD Pipeline

**Files:**
- Create: `.github/workflows/backend-ci.yml`
- Create: `.github/workflows/mobile-ci.yml`
- Create: `.github/workflows/deploy-vps.yml`

**Step 1: Create backend CI workflow**

```bash
# .github/workflows/backend-ci.yml
cat > .github/workflows/backend-ci.yml << 'EOF'
name: Backend CI

on:
  push:
    branches: [main, develop]
    paths:
      - 'backend/**'
      - '.github/workflows/backend-ci.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - 'backend/**'

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          working-directory: backend

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run tests
        working-directory: ./backend
        run: make test
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./backend/coverage.out

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec-results.sarif ./backend/...'
      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: gosec-results.sarif
EOF
```

**Step 2: Create deploy VPS workflow**

```bash
# .github/workflows/deploy-vps.yml
cat > .github/workflows/deploy-vps.yml << 'EOF'
name: Deploy to VPS

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run tests
        working-directory: ./backend
        run: make test

  deploy:
    needs: test
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4

      - name: Build and push Docker images
        run: |
          echo "Building images..."
          # TODO: Add registry login
          docker build -t iap-api:${GITHUB_SHA} -f infra/docker/backend/Dockerfile ./backend
          docker build -t iap-worker:${GITHUB_SHA} -f infra/docker/worker/Dockerfile ./backend

      - name: Deploy via SSH
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.VPS_HOST }}
          username: deploy
          key: ${{ secrets.VPS_DEPLOY_KEY }}
          script: |
            set -euo pipefail
            cd /opt/iap-system
            export IMAGE_TAG="${{ github.sha }}"
            docker compose pull
            docker compose up -d --no-deps api
            docker compose up -d --no-deps worker
EOF
```

**Step 3: Create CODEOWNERS**

```bash
# .github/CODEOWNERS
cat > .github/CODEOWNERS << 'EOF'
* @backend-team
backend/** @backend-team
mobile/** @mobile-team
infra/** @devops-team
EOF
```

**Step 4: Create PR template**

```bash
# .github/PULL_REQUEST_TEMPLATE.md
cat > .github/PULL_REQUEST_TEMPLATE.md << 'EOF'
## Description
<!-- Describe the changes in this PR -->

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Refactoring
- [ ] Documentation

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added to complex code
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] All tests passing
EOF
```

**Step 5: Commit**

```bash
git add .github/
git commit -m "chore: add CI/CD pipelines and contribution templates"
```

---

## Phase 2: Data Layer

### Task 6: Create Database Migrations

**Files:**
- Create: `backend/migrations/001_create_users.up.sql`
- Create: `backend/migrations/001_create_users.down.sql`
- Create: `backend/migrations/002_create_subscriptions.up.sql`
- Create: `backend/migrations/002_create_subscriptions.down.sql`
- Create: `backend/migrations/003_create_transactions.up.sql`
- Create: `backend/migrations/003_create_transactions.down.sql`

**Step 1: Create users table migration**

```bash
# backend/migrations/001_create_users.up.sql
cat > backend/migrations/001_create_users.up.sql << 'EOF'
-- ============================================================
-- Table: users
-- ============================================================
-- Platform identity: Apple originalTransactionId or Google
-- obfuscatedExternalAccountId. This is the canonical persistent
-- identity that survives reinstalls and device resets.
-- device_id is stored for analytics only and MUST NOT be used
-- as a foreign key.

CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_user_id    TEXT UNIQUE NOT NULL,
    device_id           TEXT,
    platform            TEXT NOT NULL CHECK (platform IN ('ios', 'android')),
    app_version         TEXT NOT NULL,
    email               TEXT UNIQUE,
    ltv                 NUMERIC(10,2) DEFAULT 0,
    ltv_updated_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);

COMMENT ON TABLE users IS 'User accounts with platform-based identity';
COMMENT ON COLUMN users.platform_user_id IS 'Canonical platform identity (Apple: originalTransactionId, Google: obfuscatedExternalAccountId)';
COMMENT ON COLUMN users.device_id IS 'Device identifier for analytics only - DO NOT use as foreign key';
COMMENT ON COLUMN users.ltv IS 'Lifetime value - updated by worker job';
EOF
```

```bash
# backend/migrations/001_create_users.down.sql
cat > backend/migrations/001_create_users.down.sql << 'EOF'
DROP TABLE IF EXISTS users CASCADE;
EOF
```

**Step 2: Create subscriptions table migration**

```bash
# backend/migrations/002_create_subscriptions.up.sql
cat > backend/migrations/002_create_subscriptions.up.sql << 'EOF'
-- ============================================================
-- Table: subscriptions
-- ============================================================

CREATE TABLE subscriptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    status          TEXT NOT NULL CHECK (status IN ('active', 'expired', 'cancelled', 'grace')),
    source          TEXT NOT NULL CHECK (source IN ('iap', 'stripe', 'paddle')),
    platform        TEXT NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    product_id      TEXT NOT NULL,
    plan_type       TEXT NOT NULL CHECK (plan_type IN ('monthly', 'annual', 'lifetime')),
    expires_at      TIMESTAMPTZ NOT NULL,
    auto_renew      BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

COMMENT ON TABLE subscriptions IS 'User subscriptions';
COMMENT ON COLUMN subscriptions.source IS 'Purchase source: iap (Apple/Google), stripe, or paddle';

-- Enforce single active subscription per user
CREATE UNIQUE INDEX idx_subscriptions_one_active
    ON subscriptions(user_id)
    WHERE status = 'active' AND deleted_at IS NULL;

-- Hot path: access check (called on every content open)
CREATE INDEX idx_subscriptions_access
    ON subscriptions(user_id, status, expires_at)
    WHERE deleted_at IS NULL;
EOF
```

```bash
# backend/migrations/002_create_subscriptions.down.sql
cat > backend/migrations/002_create_subscriptions.down.sql << 'EOF'
DROP INDEX IF EXISTS idx_subscriptions_one_active;
DROP INDEX IF EXISTS idx_subscriptions_access;
DROP TABLE IF EXISTS subscriptions CASCADE;
EOF
```

**Step 3: Create transactions table migration**

```bash
# backend/migrations/003_create_transactions.up.sql
cat > backend/migrations/003_create_transactions.up.sql << 'EOF'
-- ============================================================
-- Table: transactions
-- ============================================================
-- receipt_data stores only a SHA-256 hash of the raw receipt
-- for deduplication; the full encrypted receipt is stored as
-- a webhook_event payload.

CREATE TABLE transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id),
    subscription_id     UUID NOT NULL REFERENCES subscriptions(id),
    amount              NUMERIC(10,2) NOT NULL,
    currency            CHAR(3) NOT NULL,
    status              TEXT NOT NULL CHECK (status IN ('success', 'failed', 'refunded')),
    receipt_hash        TEXT,
    provider_tx_id      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE transactions IS 'Payment transactions';
COMMENT ON COLUMN transactions.receipt_hash IS 'SHA-256 hash of receipt for deduplication';

-- Transaction history
CREATE INDEX idx_transactions_user
    ON transactions(user_id, created_at DESC);
EOF
```

```bash
# backend/migrations/003_create_transactions.down.sql
cat > backend/migrations/003_create_transactions.down.sql << 'EOF'
DROP INDEX IF EXISTS idx_transactions_user;
DROP TABLE IF EXISTS transactions CASCADE;
EOF
```

**Step 4: Verify migration files**

```bash
ls -la backend/migrations/
```

Expected: List of all migration files created

**Step 5: Commit**

```bash
git add backend/migrations/
git commit -m "feat: add core database migrations (users, subscriptions, transactions)"
```

---

### Task 7: Create sqlc Configuration and Queries

**Files:**
- Create: `backend/internal/infrastructure/persistence/sqlc/sqlc.yaml`
- Create: `backend/internal/infrastructure/persistence/sqlc/schema.sql`
- Create: `backend/internal/infrastructure/persistence/sqlc/queries/users.sql`
- Create: `backend/internal/infrastructure/persistence/sqlc/queries/subscriptions.sql`
- Create: `backend/internal/infrastructure/persistence/sqlc/queries/transactions.sql`

**Step 1: Create sqlc.yaml**

```bash
# backend/internal/infrastructure/persistence/sqlc/sqlc.yaml
cat > backend/internal/infrastructure/persistence/sqlc/sqlc.yaml << 'EOF'
version: "2"
sql:
  - schema: "schema.sql"
    queries: "queries"
    engine: "postgresql"
    gen:
      go:
        package: "generated"
        out: "generated"
        sql_package: "github.com/jackc/pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
EOF
```

**Step 2: Create schema.sql (canonical schema from migrations)**

```bash
# backend/internal/infrastructure/persistence/sqlc/schema.sql
cat > backend/internal/infrastructure/persistence/sqlc/schema.sql << 'EOF'
-- Canonical schema for sqlc code generation
-- This file is the source of truth for sqlc
-- Keep in sync with migrations

CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform_user_id    TEXT UNIQUE NOT NULL,
    device_id           TEXT,
    platform            TEXT NOT NULL CHECK (platform IN ('ios', 'android')),
    app_version         TEXT NOT NULL,
    email               TEXT UNIQUE,
    ltv                 NUMERIC(10,2) DEFAULT 0,
    ltv_updated_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);

CREATE TABLE subscriptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    status          TEXT NOT NULL CHECK (status IN ('active', 'expired', 'cancelled', 'grace')),
    source          TEXT NOT NULL CHECK (source IN ('iap', 'stripe', 'paddle')),
    platform        TEXT NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    product_id      TEXT NOT NULL,
    plan_type       TEXT NOT NULL CHECK (plan_type IN ('monthly', 'annual', 'lifetime')),
    expires_at      TIMESTAMPTZ NOT NULL,
    auto_renew      BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id),
    subscription_id     UUID NOT NULL REFERENCES subscriptions(id),
    amount              NUMERIC(10,2) NOT NULL,
    currency            CHAR(3) NOT NULL,
    status              TEXT NOT NULL CHECK (status IN ('success', 'failed', 'refunded')),
    receipt_hash        TEXT,
    provider_tx_id      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
EOF
```

**Step 3: Create users queries**

```bash
# backend/internal/infrastructure/persistence/sqlc/queries/users.sql
cat > backend/internal/infrastructure/persistence/sqlc/queries/users.sql << 'EOF'
-- name: CreateUser :one
INSERT INTO users (platform_user_id, device_id, platform, app_version, email)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByPlatformID :one
SELECT * FROM users
WHERE platform_user_id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateUserLTV :one
UPDATE users
SET ltv = $2, ltv_updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SoftDeleteUser :one
UPDATE users
SET deleted_at = now()
WHERE id = $1
RETURNING *;
EOF
```

**Step 4: Create subscriptions queries**

```bash
# backend/internal/infrastructure/persistence/sqlc/queries/subscriptions.sql
cat > backend/internal/infrastructure/persistence/sqlc/queries/subscriptions.sql << 'EOF'
-- name: CreateSubscription :one
INSERT INTO subscriptions (user_id, status, source, platform, product_id, plan_type, expires_at, auto_renew)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetActiveSubscriptionByUserID :one
SELECT * FROM subscriptions
WHERE user_id = $1 AND status = 'active' AND deleted_at IS NULL
LIMIT 1;

-- name: GetAccessCheck :one
SELECT id, status, expires_at FROM subscriptions
WHERE user_id = $1
  AND status = 'active'
  AND expires_at > now()
  AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateSubscriptionStatus :one
UPDATE subscriptions
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateSubscriptionExpiry :one
UPDATE subscriptions
SET expires_at = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CancelSubscription :one
UPDATE subscriptions
SET status = 'cancelled', auto_renew = false, updated_at = now()
WHERE id = $1
RETURNING *;
EOF
```

**Step 5: Create transactions queries**

```bash
# backend/internal/infrastructure/persistence/sqlc/queries/transactions.sql
cat > backend/internal/infrastructure/persistence/sqlc/queries/transactions.sql << 'EOF'
-- name: CreateTransaction :one
INSERT INTO transactions (user_id, subscription_id, amount, currency, status, receipt_hash, provider_tx_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetTransactionByID :one
SELECT * FROM transactions
WHERE id = $1
LIMIT 1;

-- name: GetTransactionsByUserID :many
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CheckDuplicateReceipt :one
SELECT id FROM transactions
WHERE receipt_hash = $1
LIMIT 1;
EOF
```

**Step 6: Verify sqlc configuration**

```bash
cd backend/internal/infrastructure/persistence/sqlc
ls -la
```

Expected: Directory listing with sqlc.yaml, schema.sql, queries/ directory

**Step 7: Commit**

```bash
git add backend/internal/infrastructure/persistence/sqlc/
git commit -m "feat: add sqlc configuration and core queries"
```

---

### Task 8: Set Up Domain Layer - Entities

**Files:**
- Create: `backend/internal/domain/entity/user.go`
- Create: `backend/internal/domain/entity/subscription.go`
- Create: `backend/internal/domain/entity/transaction.go`

**Step 1: Create user entity**

```bash
# backend/internal/domain/entity/user.go
cat > backend/internal/domain/entity/user.go << 'EOF'
package entity

import (
	"time"

	"github.com/google/uuid"
)

type Platform string

const (
	PlatformiOS     Platform = "ios"
	PlatformAndroid Platform = "android"
)

type User struct {
	ID            uuid.UUID
	PlatformUserID string
	DeviceID      string
	Platform      Platform
	AppVersion    string
	Email         string
	LTV           float64
	LTVUpdatedAt  time.Time
	CreatedAt     time.Time
	DeletedAt     *time.Time
}

// NewUser creates a new user entity
func NewUser(platformUserID, deviceID string, platform Platform, appVersion, email string) *User {
	return &User{
		ID:            uuid.New(),
		PlatformUserID: platformUserID,
		DeviceID:      deviceID,
		Platform:      platform,
		AppVersion:    appVersion,
		Email:         email,
		LTV:           0,
		CreatedAt:     time.Now(),
	}
}

// IsDeleted returns true if the user has been soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// HasEmail returns true if the user has an email address
func (u *User) HasEmail() bool {
	return u.Email != ""
}
EOF
```

**Step 2: Create subscription entity**

```bash
# backend/internal/domain/entity/subscription.go
cat > backend/internal/domain/entity/subscription.go << 'EOF'
package entity

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusExpired   SubscriptionStatus = "expired"
	StatusCancelled SubscriptionStatus = "cancelled"
	StatusGrace     SubscriptionStatus = "grace"
)

type SubscriptionSource string

const (
	SourceIAP    SubscriptionSource = "iap"
	SourceStripe SubscriptionSource = "stripe"
	SourcePaddle SubscriptionSource = "paddle"
)

type PlanType string

const (
	PlanMonthly  PlanType = "monthly"
	PlanAnnual   PlanType = "annual"
	PlanLifetime PlanType = "lifetime"
)

type Subscription struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Status        SubscriptionStatus
	Source        SubscriptionSource
	Platform      string
	ProductID     string
	PlanType      PlanType
	ExpiresAt     time.Time
	AutoRenew     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// NewSubscription creates a new subscription entity
func NewSubscription(userID uuid.UUID, source SubscriptionSource, platform, productID string, planType PlanType, expiresAt time.Time) *Subscription {
	return &Subscription{
		ID:        uuid.New(),
		UserID:    userID,
		Status:    StatusActive,
		Source:    source,
		Platform:  platform,
		ProductID: productID,
		PlanType:  planType,
		ExpiresAt: expiresAt,
		AutoRenew: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// IsActive returns true if the subscription is currently active
func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive && s.ExpiresAt.After(time.Now())
}

// IsExpired returns true if the subscription has expired
func (s *Subscription) IsExpired() bool {
	return s.Status == StatusExpired || s.ExpiresAt.Before(time.Now())
}

// CanAccessContent returns true if user can access premium content
func (s *Subscription) CanAccessContent() bool {
	if s.DeletedAt != nil {
		return false
	}
	return s.IsActive()
}

// HasGracePeriod returns true if subscription is in grace period
func (s *Subscription) HasGracePeriod() bool {
	return s.Status == StatusGrace
}
EOF
```

**Step 3: Create transaction entity**

```bash
# backend/internal/domain/entity/transaction.go
cat > backend/internal/domain/entity/transaction.go << 'EOF'
package entity

import (
	"time"

	"github.com/google/uuid"
)

type TransactionStatus string

const (
	TransactionStatusSuccess  TransactionStatus = "success"
	TransactionStatusFailed   TransactionStatus = "failed"
	TransactionStatusRefunded TransactionStatus = "refunded"
)

type Transaction struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	SubscriptionID uuid.UUID
	Amount         float64
	Currency       string
	Status         TransactionStatus
	ReceiptHash    string
	ProviderTxID   string
	CreatedAt      time.Time
}

// NewTransaction creates a new transaction entity
func NewTransaction(userID, subscriptionID uuid.UUID, amount float64, currency string) *Transaction {
	return &Transaction{
		ID:             uuid.New(),
		UserID:         userID,
		SubscriptionID: subscriptionID,
		Amount:         amount,
		Currency:       currency,
		Status:         TransactionStatusSuccess,
		CreatedAt:      time.Now(),
	}
}

// IsSuccessful returns true if the transaction was successful
func (t *Transaction) IsSuccessful() bool {
	return t.Status == TransactionStatusSuccess
}

// IsFailed returns true if the transaction failed
func (t *Transaction) IsFailed() bool {
	return t.Status == TransactionStatusFailed
}
EOF
```

**Step 4: Verify entities compile**

```bash
cd backend
go build ./internal/domain/entity/...
```

Expected: No errors

**Step 5: Commit**

```bash
git add backend/internal/domain/entity/
git commit -m "feat: add core domain entities (User, Subscription, Transaction)"
```

---

### Task 9: Set Up Domain Layer - Value Objects

**Files:**
- Create: `backend/internal/domain/valueobject/email.go`
- Create: `backend/internal/domain/valueobject/money.go`
- Create: `backend/internal/domain/valueobject/plan_type.go`
- Create: `backend/internal/domain/valueobject/subscription_status.go`

**Step 1: Create email value object**

```bash
# backend/internal/domain/valueobject/email.go
cat > backend/internal/domain/valueobject/email.go << 'EOF'
package valueobject

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
)

// Email represents a valid email address
type Email struct {
	value string
}

// NewEmail creates a new Email value object
func NewEmail(email string) (*Email, error) {
	if !isValidEmail(email) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidEmail, email)
	}
	return &Email{value: email}, nil
}

// String returns the email string
func (e *Email) String() string {
	return e.value
}

// isValidEmail checks if the email format is valid
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	// Simple email regex for validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
EOF
```

**Step 2: Create money value object**

```bash
# backend/internal/domain/valueobject/money.go
cat > backend/internal/domain/valueobject/money.go << 'EOF'
package valueobject

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidAmount   = errors.New("amount must be non-negative")
	ErrInvalidCurrency = errors.New("invalid currency code")
)

// Money represents a monetary value
type Money struct {
	Amount   float64
	Currency string // ISO 4217 currency code (e.g., "USD", "EUR")
}

// NewMoney creates a new Money value object
func NewMoney(amount float64, currency string) (*Money, error) {
	if amount < 0 {
		return nil, fmt.Errorf("%w: %f", ErrInvalidAmount, amount)
	}
	if !isValidCurrency(currency) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCurrency, currency)
	}
	return &Money{
		Amount:   amount,
		Currency: currency,
	}, nil
}

// isValidCurrency checks if the currency code is valid (3 letters)
func isValidCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, c := range currency {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			return false
		}
	}
	return true
}

// String returns a string representation of the money
func (m *Money) String() string {
	return fmt.Sprintf("%.2f %s", m.Amount, m.Currency)
}

// IsZero returns true if the amount is zero
func (m *Money) IsZero() bool {
	return m.Amount == 0
}

// Add adds another Money value to this one
func (m *Money) Add(other *Money) (*Money, error) {
	if m.Currency != other.Currency {
		return nil, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
	}
	return NewMoney(m.Amount+other.Amount, m.Currency)
}
EOF
```

**Step 3: Create plan type value object**

```bash
# backend/internal/domain/valueobject/plan_type.go
cat > backend/internal/domain/valueobject/plan_type.go << 'EOF'
package valueobject

import (
	"errors"
)

var (
	ErrInvalidPlanType = errors.New("invalid plan type")
)

type PlanType string

const (
	PlanMonthly  PlanType = "monthly"
	PlanAnnual   PlanType = "annual"
	PlanLifetime PlanType = "lifetime"
)

// NewPlanType creates a new PlanType value object
func NewPlanType(planType string) (PlanType, error) {
	pt := PlanType(planType)
	switch pt {
	case PlanMonthly, PlanAnnual, PlanLifetime:
		return pt, nil
	default:
		return "", ErrInvalidPlanType
	}
}

// String returns the string representation of the plan type
func (p PlanType) String() string {
	return string(p)
}

// IsValid returns true if the plan type is valid
func (p PlanType) IsValid() bool {
	switch p {
	case PlanMonthly, PlanAnnual, PlanLifetime:
		return true
	default:
		return false
	}
}
EOF
```

**Step 4: Create subscription status value object**

```bash
# backend/internal/domain/valueobject/subscription_status.go
cat > backend/internal/domain/valueobject/subscription_status.go << 'EOF'
package valueobject

import (
	"errors"
)

var (
	ErrInvalidSubscriptionStatus = errors.New("invalid subscription status")
)

type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusExpired   SubscriptionStatus = "expired"
	StatusCancelled SubscriptionStatus = "cancelled"
	StatusGrace     SubscriptionStatus = "grace"
)

// NewSubscriptionStatus creates a new SubscriptionStatus value object
func NewSubscriptionStatus(status string) (SubscriptionStatus, error) {
	s := SubscriptionStatus(status)
	switch s {
	case StatusActive, StatusExpired, StatusCancelled, StatusGrace:
		return s, nil
	default:
		return "", ErrInvalidSubscriptionStatus
	}
}

// String returns the string representation of the status
func (s SubscriptionStatus) String() string {
	return string(s)
}

// IsActive returns true if the status is active
func (s SubscriptionStatus) IsActive() bool {
	return s == StatusActive
}

// IsTerminated returns true if the subscription is cancelled or expired
func (s SubscriptionStatus) IsTerminated() bool {
	return s == StatusCancelled || s == StatusExpired
}
EOF
```

**Step 5: Verify value objects compile**

```bash
cd backend
go build ./internal/domain/valueobject/...
```

Expected: No errors

**Step 6: Commit**

```bash
git add backend/internal/domain/valueobject/
git commit -m "feat: add domain value objects (Email, Money, PlanType, SubscriptionStatus)"
```

---

### Task 10: Set Up Domain Layer - Repository Interfaces

**Files:**
- Create: `backend/internal/domain/repository/user_repository.go`
- Create: `backend/internal/domain/repository/subscription_repository.go`
- Create: `backend/internal/domain/repository/transaction_repository.go`

**Step 1: Create user repository interface**

```bash
# backend/internal/domain/repository/user_repository.go
cat > backend/internal/domain/repository/user_repository.go << 'EOF'
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/your-org/iap-system/internal/domain/entity"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetByPlatformID retrieves a user by platform user ID
	GetByPlatformID(ctx context.Context, platformUserID string) (*entity.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *entity.User) error

	// SoftDelete soft deletes a user
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// ExistsByPlatformID checks if a user exists with the given platform ID
	ExistsByPlatformID(ctx context.Context, platformUserID string) (bool, error)
}
EOF
```

**Step 2: Create subscription repository interface**

```bash
# backend/internal/domain/repository/subscription_repository.go
cat > backend/internal/domain/repository/subscription_repository.go << 'EOF'
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/your-org/iap-system/internal/domain/entity"
)

// SubscriptionRepository defines the interface for subscription data access
type SubscriptionRepository interface {
	// Create creates a new subscription
	Create(ctx context.Context, subscription *entity.Subscription) error

	// GetByID retrieves a subscription by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)

	// GetActiveByUserID retrieves the active subscription for a user
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error)

	// GetByUserID retrieves all subscriptions for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Subscription, error)

	// Update updates an existing subscription
	Update(ctx context.Context, subscription *entity.Subscription) error

	// UpdateStatus updates the status of a subscription
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SubscriptionStatus) error

	// UpdateExpiry updates the expiry date of a subscription
	UpdateExpiry(ctx context.Context, id uuid.UUID, expiresAt interface{}) error

	// Cancel cancels a subscription
	Cancel(ctx context.Context, id uuid.UUID) error

	// CanAccess checks if a user can access premium content
	CanAccess(ctx context.Context, userID uuid.UUID) (bool, error)
}
EOF
```

**Step 3: Create transaction repository interface**

```bash
# backend/internal/domain/repository/transaction_repository.go
cat > backend/internal/domain/repository/transaction_repository.go << 'EOF'
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/your-org/iap-system/internal/domain/entity"
)

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	// Create creates a new transaction
	Create(ctx context.Context, transaction *entity.Transaction) error

	// GetByID retrieves a transaction by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)

	// GetByUserID retrieves transactions for a user with pagination
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Transaction, error)

	// GetBySubscriptionID retrieves transactions for a subscription
	GetBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entity.Transaction, error)

	// CheckDuplicateReceipt checks if a receipt has already been processed
	CheckDuplicateReceipt(ctx context.Context, receiptHash string) (bool, error)
}
EOF
```

**Step 4: Verify repository interfaces compile**

```bash
cd backend
go build ./internal/domain/repository/...
```

Expected: No errors

**Step 5: Commit**

```bash
git add backend/internal/domain/repository/
git commit -m "feat: add domain repository interfaces"
```

---

### Task 11: Set Up Domain Layer - Errors

**Files:**
- Create: `backend/internal/domain/errors/domain_errors.go`
- Create: `backend/internal/domain/errors/validation_errors.go`

**Step 1: Create domain errors**

```bash
# backend/internal/domain/errors/domain_errors.go
cat > backend/internal/domain/errors/domain_errors.go << 'EOF'
package errors

import (
	"errors"
	"fmt"
)

var (
	// User errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	// Subscription errors
	ErrSubscriptionNotFound    = errors.New("subscription not found")
	ErrSubscriptionNotActive   = errors.New("subscription is not active")
	ErrSubscriptionExpired     = errors.New("subscription has expired")
	ErrSubscriptionCancelled   = errors.New("subscription has been cancelled")
	ErrActiveSubscriptionExists = errors.New("active subscription already exists")

	// Transaction errors
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrDuplicateReceipt       = errors.New("receipt has already been processed")
	ErrReceiptInvalid         = errors.New("receipt is invalid")
	ErrReceiptExpired         = errors.New("receipt has expired")

	// Payment errors
	ErrPaymentFailed   = errors.New("payment failed")
	ErrPaymentRefunded = errors.New("payment has been refunded")

	// External service errors
	ErrExternalServiceUnavailable = errors.New("external service unavailable")
	ErrIAPVerificationFailed     = errors.New("IAP verification failed")
)

// NotFoundError wraps an error with not found context
type NotFoundError struct {
	Entity string
	ID     string
	Err    error
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id '%s' not found: %v", e.Entity, e.ID, e.Err)
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

// ConflictError wraps an error with conflict context
type ConflictError struct {
	Entity string
	Reason string
	Err    error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s conflict: %s - %v", e.Entity, e.Reason, e.Err)
}

func (e *ConflictError) Unwrap() error {
	return e.Err
}
EOF
```

**Step 2: Create validation errors**

```bash
# backend/internal/domain/errors/validation_errors.go
cat > backend/internal/domain/errors/validation_errors.go << 'EOF'
package errors

import (
	"errors"
	"fmt"
)

var (
	// General validation errors
	ErrInvalidInput   = errors.New("invalid input")
	ErrRequiredField  = errors.New("required field is missing")
	ErrInvalidFormat  = errors.New("invalid format")
	ErrInvalidLength  = errors.New("invalid length")
	ErrOutOfRange     = errors.New("value out of range")

	// Specific field validation errors
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrInvalidPlatform    = errors.New("invalid platform")
	ErrInvalidPlanType    = errors.New("invalid plan type")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidAmount      = errors.New("invalid amount")
	ErrInvalidCurrency    = errors.New("invalid currency code")
	ErrInvalidReceipt     = errors.New("invalid receipt data")
)

// ValidationError wraps a field validation error
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation failed for field '%s': %v", e.Field, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
EOF
```

**Step 3: Verify errors compile**

```bash
cd backend
go build ./internal/domain/errors/...
```

Expected: No errors

**Step 4: Commit**

```bash
git add backend/internal/domain/errors/
git commit -m "feat: add domain and validation error types"
```

---

### Task 12: Implement Infrastructure - Database Connection Pool

**Files:**
- Create: `backend/internal/infrastructure/persistence/pool/pgxpool.go`
- Create: `backend/internal/infrastructure/config/database_config.go`

**Step 1: Create database config**

```bash
# backend/internal/infrastructure/config/database_config.go
cat > backend/internal/infrastructure/config/database_config.go << 'EOF'
package config

import (
	"time"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL            string
	MaxConnections int
	MinConnections int
	MaxLifetime    time.Duration
	MaxIdleTime    time.Duration
	HealthCheck    time.Duration
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		MaxConnections: 25,
		MinConnections: 5,
		MaxLifetime:    1 * time.Hour,
		MaxIdleTime:    30 * time.Minute,
		HealthCheck:    30 * time.Second,
	}
}
EOF
```

**Step 2: Create connection pool**

```bash
# backend/internal/infrastructure/persistence/pool/pgxpool.go
cat > backend/internal/infrastructure/persistence/pool/pgxpool.go << 'EOF'
package pool

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/iap-system/internal/infrastructure/config"
)

// NewPool creates a new PostgreSQL connection pool
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Max connections: sized at 25 per API instance.
	// With 2 API replicas + 1 worker = 75 total.
	// Postgres default max_connections = 100 → leaves 25 for migrations/admin.
	config.MaxConns = int32(cfg.MaxConnections)
	config.MinConns = int32(cfg.MinConnections)
	config.MaxConnLifetime = cfg.MaxLifetime
	config.MaxConnIdleTime = cfg.MaxIdleTime
	config.HealthCheckPeriod = cfg.HealthCheck

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pool, nil
}

// Ping verifies the database connection is alive
func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}

// Close closes the connection pool
func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}
EOF
```

**Step 3: Verify pool code compiles**

```bash
cd backend
go build ./internal/infrastructure/persistence/pool/...
```

Expected: No errors

**Step 4: Commit**

```bash
git add backend/internal/infrastructure/persistence/pool/ backend/internal/infrastructure/config/
git commit -m "feat: add database connection pool configuration"
```

---

### Task 13: Implement Infrastructure - Repository Implementations

**Files:**
- Create: `backend/internal/infrastructure/persistence/repository/user_repository_impl.go`
- Create: `backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go`
- Create: `backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go`

**Step 1: Create user repository implementation**

```bash
# backend/internal/infrastructure/persistence/repository/user_repository_impl.go
cat > backend/internal/infrastructure/persistence/repository/user_repository_impl.go << 'EOF'
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/your-org/iap-system/internal/domain/entity"
	"github.com/your-org/iap-system/internal/domain/repository"
	domainErrors "github.com/your-org/iap-system/internal/domain/errors"
	"github.com/your-org/iap-system/internal/infrastructure/persistence/sqlc/generated"
)

type userRepositoryImpl struct {
	queries *generated.Queries
}

// NewUserRepository creates a new user repository implementation
func NewUserRepository(queries *generated.Queries) repository.UserRepository {
	return &userRepositoryImpl{queries: queries}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	params := generated.CreateUserParams{
		PlatformUserID: user.PlatformUserID,
		DeviceID:      user.DeviceID,
		Platform:      string(user.Platform),
		AppVersion:    user.AppVersion,
		Email:         user.Email,
	}

	_, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", domainErrors.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *userRepositoryImpl) GetByPlatformID(ctx context.Context, platformUserID string) (*entity.User, error) {
	row, err := r.queries.GetUserByPlatformID(ctx, platformUserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", domainErrors.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", domainErrors.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *userRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	params := generated.UpdateUserLTVParams{
		ID:  user.ID,
		Ltv: user.LTV,
	}

	_, err := r.queries.UpdateUserLTV(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) ExistsByPlatformID(ctx context.Context, platformUserID string) (bool, error) {
	_, err := r.queries.GetUserByPlatformID(ctx, platformUserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return true, nil
}

func (r *userRepositoryImpl) mapToEntity(row generated.User) *entity.User {
	return &entity.User{
		ID:            row.ID,
		PlatformUserID: row.PlatformUserID,
		DeviceID:      row.DeviceID,
		Platform:      entity.Platform(row.Platform),
		AppVersion:    row.AppVersion,
		Email:         row.Email,
		LTV:           row.Ltv,
		LTVUpdatedAt:  row.LtvUpdatedAt.Time,
		CreatedAt:     row.CreatedAt,
		DeletedAt:     row.DeletedAt,
	}
}
EOF
```

**Step 2: Create subscription repository implementation**

```bash
# backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go
cat > backend/internal/infrastructure/persistence/repository/subscription_repository_impl.go << 'EOF'
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/your-org/iap-system/internal/domain/entity"
	"github.com/your-org/iap-system/internal/domain/repository"
	domainErrors "github.com/your-org/iap-system/internal/domain/errors"
	"github.com/your-org/iap-system/internal/infrastructure/persistence/sqlc/generated"
)

type subscriptionRepositoryImpl struct {
	queries *generated.Queries
}

// NewSubscriptionRepository creates a new subscription repository implementation
func NewSubscriptionRepository(queries *generated.Queries) repository.SubscriptionRepository {
	return &subscriptionRepositoryImpl{queries: queries}
}

func (r *subscriptionRepositoryImpl) Create(ctx context.Context, sub *entity.Subscription) error {
	params := generated.CreateSubscriptionParams{
		UserID:    sub.UserID,
		Status:    string(sub.Status),
		Source:    string(sub.Source),
		Platform:  sub.Platform,
		ProductID: sub.ProductID,
		PlanType:  string(sub.PlanType),
		ExpiresAt: sub.ExpiresAt,
		AutoRenew: sub.AutoRenew,
	}

	_, err := r.queries.CreateSubscription(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (r *subscriptionRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	row, err := r.queries.GetSubscriptionByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("subscription not found: %w", domainErrors.ErrSubscriptionNotFound)
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *subscriptionRepositoryImpl) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*entity.Subscription, error) {
	row, err := r.queries.GetActiveSubscriptionByUserID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("active subscription not found: %w", domainErrors.ErrSubscriptionNotActive)
		}
		return nil, fmt.Errorf("failed to get active subscription: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *subscriptionRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Subscription, error) {
	// For now, return empty array if no subscriptions
	// In full implementation, add query to GetSubscriptionsByUserID
	return []*entity.Subscription{}, nil
}

func (r *subscriptionRepositoryImpl) Update(ctx context.Context, sub *entity.Subscription) error {
	params := generated.UpdateSubscriptionStatusParams{
		ID:     sub.ID,
		Status: string(sub.Status),
	}

	_, err := r.queries.UpdateSubscriptionStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (r *subscriptionRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SubscriptionStatus) error {
	params := generated.UpdateSubscriptionStatusParams{
		ID:     id,
		Status: string(status),
	}

	_, err := r.queries.UpdateSubscriptionStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	return nil
}

func (r *subscriptionRepositoryImpl) UpdateExpiry(ctx context.Context, id uuid.UUID, expiresAt interface{}) error {
	params := generated.UpdateSubscriptionExpiryParams{
		ID:        id,
		ExpiresAt: expiresAt.(time.Time),
	}

	_, err := r.queries.UpdateSubscriptionExpiry(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update subscription expiry: %w", err)
	}

	return nil
}

func (r *subscriptionRepositoryImpl) Cancel(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.CancelSubscription(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}

func (r *subscriptionRepositoryImpl) CanAccess(ctx context.Context, userID uuid.UUID) (bool, error) {
	_, err := r.queries.GetAccessCheck(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check access: %w", err)
	}

	return true, nil
}

func (r *subscriptionRepositoryImpl) mapToEntity(row generated.Subscription) *entity.Subscription {
	return &entity.Subscription{
		ID:        row.ID,
		UserID:    row.UserID,
		Status:    entity.SubscriptionStatus(row.Status),
		Source:    entity.SubscriptionSource(row.Source),
		Platform:  row.Platform,
		ProductID: row.ProductID,
		PlanType:  entity.PlanType(row.PlanType),
		ExpiresAt: row.ExpiresAt,
		AutoRenew: row.AutoRenew,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: row.DeletedAt,
	}
}
EOF
```

**Step 3: Create transaction repository implementation**

```bash
# backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go
cat > backend/internal/infrastructure/persistence/repository/transaction_repository_impl.go << 'EOF'
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/your-org/iap-system/internal/domain/entity"
	"github.com/your-org/iap-system/internal/domain/repository"
	domainErrors "github.com/your-org/iap-system/internal/domain/errors"
	"github.com/your-org/iap-system/internal/infrastructure/persistence/sqlc/generated"
)

type transactionRepositoryImpl struct {
	queries *generated.Queries
}

// NewTransactionRepository creates a new transaction repository implementation
func NewTransactionRepository(queries *generated.Queries) repository.TransactionRepository {
	return &transactionRepositoryImpl{queries: queries}
}

func (r *transactionRepositoryImpl) Create(ctx context.Context, txn *entity.Transaction) error {
	params := generated.CreateTransactionParams{
		UserID:         txn.UserID,
		SubscriptionID: txn.SubscriptionID,
		Amount:         txn.Amount,
		Currency:       txn.Currency,
		Status:         string(txn.Status),
		ReceiptHash:    txn.ReceiptHash,
		ProviderTxID:   txn.ProviderTxID,
	}

	_, err := r.queries.CreateTransaction(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *transactionRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	row, err := r.queries.GetTransactionByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %w", domainErrors.ErrTransactionNotFound)
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return r.mapToEntity(row), nil
}

func (r *transactionRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Transaction, error) {
	params := generated.GetTransactionsByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	rows, err := r.queries.GetTransactionsByUserID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	transactions := make([]*entity.Transaction, len(rows))
	for i, row := range rows {
		transactions[i] = r.mapToEntity(row)
	}

	return transactions, nil
}

func (r *transactionRepositoryImpl) GetBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entity.Transaction, error) {
	// For now, return empty array
	// In full implementation, add query to GetTransactionsBySubscriptionID
	return []*entity.Transaction{}, nil
}

func (r *transactionRepositoryImpl) CheckDuplicateReceipt(ctx context.Context, receiptHash string) (bool, error) {
	_, err := r.queries.CheckDuplicateReceipt(ctx, receiptHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check duplicate receipt: %w", err)
	}

	return true, nil
}

func (r *transactionRepositoryImpl) mapToEntity(row generated.Transaction) *entity.Transaction {
	return &entity.Transaction{
		ID:             row.ID,
		UserID:         row.UserID,
		SubscriptionID: row.SubscriptionID,
		Amount:         row.Amount,
		Currency:       row.Currency,
		Status:         entity.TransactionStatus(row.Status),
		ReceiptHash:    row.ReceiptHash,
		ProviderTxID:   row.ProviderTxID,
		CreatedAt:      row.CreatedAt,
	}
}
EOF
```

**Step 4: Verify repository implementations compile**

```bash
cd backend
go build ./internal/infrastructure/persistence/repository/...
```

Expected: No errors (sqlc generated code not created yet, but interface should match)

**Step 5: Commit**

```bash
git add backend/internal/infrastructure/persistence/repository/
git commit -m "feat: add repository implementations for users, subscriptions, transactions"
```

---

### Task 14: Generate sqlc Code

**Files:**
- Modify: `backend/go.mod`
- Create: `backend/scripts/sqlc-generate.sh`

**Step 1: Add sqlc to go.mod**

```bash
cd backend
go get github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

**Step 2: Create sqlc generate script**

```bash
# backend/scripts/sqlc-generate.sh
cat > scripts/sqlc-generate.sh << 'EOF'
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../internal/infrastructure/persistence/sqlc"

echo "Generating sqlc code..."
sqlc generate

echo "sqlc code generated successfully"
EOF
chmod +x scripts/sqlc-generate.sh
```

**Step 3: Install sqlc**

```bash
# Install sqlc if not already installed
if ! command -v sqlc &> /dev/null; then
    echo "Installing sqlc..."
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
fi
```

**Step 4: Generate sqlc code**

```bash
cd backend
./scripts/sqlc-generate.sh
```

Expected: sqlc generates Go code in `generated/` directory

**Step 5: Verify generated code**

```bash
ls -la backend/internal/infrastructure/persistence/sqlc/generated/
```

Expected: db.go, models.go, and *.sql.go files

**Step 6: Commit**

```bash
git add backend/internal/infrastructure/persistence/sqlc/generated/ backend/scripts/
git commit -m "chore: generate sqlc code from database schema and queries"
```

---

## Summary and Next Steps

This implementation plan covers:

1. **Phase 1: Project Setup**
   - Repository structure and configuration files
   - Backend directory structure (clean architecture)
   - Go module initialization
   - Docker Compose configurations
   - CI/CD pipelines

2. **Phase 2: Data Layer**
   - Database migrations (users, subscriptions, transactions)
   - sqlc configuration and queries
   - Domain entities and value objects
   - Repository interfaces and implementations
   - Connection pool configuration

**Remaining Phases** (to be added in continuation):

- **Phase 3: Backend Core** - Auth, JWT, IAP verification, webhooks
- **Phase 4: Mobile App** - React Native setup, react-native-iap integration
- **Phase 5: Integration** - E2E tests, load tests
- **Phase 6: Production** - Grace periods, winback, A/B testing
- **Phase 7: Security** - GDPR endpoints, penetration testing

---

**Plan complete and saved to `docs/plans/2026-02-28-iap-system-implementation.md`.**

**Two execution options:**

1. **Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration
2. **Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
