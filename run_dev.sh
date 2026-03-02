#!/usr/bin/env bash
# run_dev.sh — cold-start the full paywall-iap dev stack
#
# What it does:
#   1. Ensures Docker Desktop is running
#   2. Tears down any stale containers (orphans too)
#   3. Detects port conflicts and picks a free API port
#   4. Starts postgres + redis + api + worker (docker-compose.local.yml)
#   5. Runs migrations, auto-fixes known dirty-flag issues
#   6. Seeds superadmin if not present
#   7. Starts Next.js frontend with hot-reload (docker-compose.dev.yml)
#   8. Prints health summary + URLs
#
# Usage:
#   ./run_dev.sh                         # defaults: admin@paywall.local / admin12345
#   ./run_dev.sh stop                    # stop everything
#   ./run_dev.sh logs                    # tail all logs
#   ADMIN_EMAIL=me@x.com ADMIN_PASS=secret ./run_dev.sh

set -euo pipefail

# ─── Config ──────────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_COMPOSE="$SCRIPT_DIR/infra/docker-compose/docker-compose.local.yml"
FRONTEND_COMPOSE="$SCRIPT_DIR/frontend/docker-compose.dev.yml"
MOCK_COMPOSE="$SCRIPT_DIR/tests/google-billing-mock/deploy/docker-compose.yml"
DB_CONTAINER="docker-compose-db-1"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@paywall.local}"
ADMIN_PASS="${ADMIN_PASS:-admin12345}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"
API_PORT_INTERNAL=8080
API_PORT_HOST="${API_PORT_HOST:-}"   # auto-detected if empty

# ─── Colors ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
info()    { echo -e "${CYAN}▶ $*${NC}"; }
ok()      { echo -e "${GREEN}✔ $*${NC}"; }
warn()    { echo -e "${YELLOW}⚠ $*${NC}"; }
die()     { echo -e "${RED}✖ $*${NC}" >&2; exit 1; }

# ─── Helpers ─────────────────────────────────────────────────────────────────
port_free() { ! lsof -i ":$1" -sTCP:LISTEN -t &>/dev/null; }

wait_for_healthy() {
  local container="$1" max="${2:-60}" i=0
  info "Waiting for $container to be healthy..."
  until [[ "$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null)" == "healthy" ]]; do
    ((i++))
    [[ $i -ge $max ]] && die "$container did not become healthy after ${max}s"
    sleep 1
  done
  ok "$container is healthy"
}

wait_for_http() {
  local url="$1" label="${2:-$1}" max="${3:-60}" i=0
  info "Waiting for $label..."
  until curl -sf -L -o /dev/null "$url"; do
    ((i++))
    [[ $i -ge $max ]] && die "$label not reachable after ${max}s"
    sleep 1
  done
  ok "$label is up"
}

# ─── Sub-commands ─────────────────────────────────────────────────────────────
cmd_stop() {
  info "Stopping all containers..."
  docker compose -f "$BACKEND_COMPOSE"  down --remove-orphans 2>/dev/null || true
  docker compose -f "$FRONTEND_COMPOSE" down --remove-orphans 2>/dev/null || true
  docker compose -f "$MOCK_COMPOSE"     down --remove-orphans 2>/dev/null || true
  ok "All stopped"
}

cmd_logs() {
  docker compose -f "$BACKEND_COMPOSE"  logs -f --tail=50 &
  docker compose -f "$FRONTEND_COMPOSE" logs -f --tail=50
}

# ─── Main ─────────────────────────────────────────────────────────────────────
[[ "${1:-}" == "stop" ]] && { cmd_stop; exit 0; }
[[ "${1:-}" == "logs" ]] && { cmd_logs; exit 0; }

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════╗"
echo -e "║   paywall-iap  —  dev cold start         ║"
echo -e "╚══════════════════════════════════════════╝${NC}"
echo ""

# ── Step 1: Docker daemon ─────────────────────────────────────────────────────
info "Checking Docker daemon..."
if ! docker info &>/dev/null; then
  warn "Docker not running — launching Docker Desktop..."
  open -a Docker
  echo -n "  Waiting for daemon"
  until docker info &>/dev/null; do echo -n "."; sleep 2; done
  echo ""
fi
ok "Docker is running"

# ── Step 2: Tear down stale containers ───────────────────────────────────────
info "Cleaning up stale containers..."
docker compose -f "$BACKEND_COMPOSE"  down --remove-orphans -t 5 2>/dev/null || true
docker compose -f "$FRONTEND_COMPOSE" down --remove-orphans -t 5 2>/dev/null || true
ok "Cleanup done"

# ── Step 3: Pick free API host port ──────────────────────────────────────────
if [[ -z "$API_PORT_HOST" ]]; then
  for try_port in 8080 8081 8082 8083; do
    if port_free $try_port; then
      API_PORT_HOST=$try_port
      break
    fi
    warn "Port $try_port in use, trying next..."
  done
  [[ -z "$API_PORT_HOST" ]] && die "No free port found in range 8080-8083"
fi

# Patch port mapping in local compose (sed in-place, keeps original logic)
info "Using API host port: $API_PORT_HOST → container $API_PORT_INTERNAL"
sed -i.bak \
  "s|\"[0-9]*:${API_PORT_INTERNAL}\"|\"${API_PORT_HOST}:${API_PORT_INTERNAL}\"|g" \
  "$BACKEND_COMPOSE"
rm -f "${BACKEND_COMPOSE}.bak"

# ── Step 4: Start backend (DB, Redis, API, Worker) ───────────────────────────
info "Starting backend stack..."
docker compose -f "$BACKEND_COMPOSE" up -d --build

wait_for_healthy "$DB_CONTAINER" 60
ok "Backend stack up"

# ── Step 5: Run migrations (with auto-fix for known SQL issues) ───────────────
info "Running migrations..."

# Auto-fix: remove now() from partial index predicates (not IMMUTABLE)
sed -i.bak \
  "s/WHERE status = 'pending' AND next_retry_at <= now()/WHERE status = 'pending'/g;
   s/WHERE status = 'sent' AND sent_at < NOW() - INTERVAL '30 days'/WHERE status = 'sent'/g" \
  "$SCRIPT_DIR/backend/migrations/016_create_matomo_staged_events.up.sql" 2>/dev/null || true
rm -f "$SCRIPT_DIR/backend/migrations/016_create_matomo_staged_events.up.sql.bak"

# Auto-fix: partial UNIQUE constraint inside CREATE TABLE (not supported, needs separate index)
sed -i.bak \
  "s/CONSTRAINT unique_active_assignment UNIQUE (experiment_id, user_id)[[:space:]]*WHERE expires_at > now()/CONSTRAINT unique_assignment UNIQUE (experiment_id, user_id)/g" \
  "$SCRIPT_DIR/backend/migrations/015_create_ab_test_assignments.up.sql" 2>/dev/null || true
rm -f "$SCRIPT_DIR/backend/migrations/015_create_ab_test_assignments.up.sql.bak"

# Clear any dirty flag before running migrator
docker exec "$DB_CONTAINER" psql -U postgres -d iap_db \
  -c "UPDATE schema_migrations SET dirty = false WHERE dirty = true;" &>/dev/null || true

# Build + run migrator
docker compose -f "$BACKEND_COMPOSE" build migrator 2>&1 | tail -3
docker compose -f "$BACKEND_COMPOSE" run --rm migrator 2>&1 | tail -5

ok "Migrations done"

# ── Step 6: Seed superadmin ────────────────────────────────────────────────────
info "Checking for superadmin..."
EXISTING=$(docker exec "$DB_CONTAINER" psql -U postgres -d iap_db \
  -tAc "SELECT COUNT(*) FROM users WHERE role='superadmin';" 2>/dev/null || echo "0")
EXISTING="${EXISTING//[[:space:]]/}"

if [[ "$EXISTING" == "0" ]]; then
  info "Seeding admin: $ADMIN_EMAIL"
  DB_CONTAINER="$DB_CONTAINER" bash "$SCRIPT_DIR/scripts/seed_admin.sh" "$ADMIN_EMAIL" "$ADMIN_PASS"
else
  ok "Superadmin already exists ($EXISTING found) — skipping seed"
fi

# ── Step 7: Start frontend ─────────────────────────────────────────────────────
info "Starting frontend (hot-reload)..."
docker compose -f "$FRONTEND_COMPOSE" up -d --build
ok "Frontend container started"

# ── Step 8: Health checks ──────────────────────────────────────────────────────
wait_for_http "http://localhost:${API_PORT_HOST}/health" "Backend API" 30
wait_for_http "http://localhost:${FRONTEND_PORT}"        "Frontend"    90

# ── Done ───────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════╗"
echo -e "║  ✅  Dev stack is ready!                             ║"
echo -e "╠══════════════════════════════════════════════════════╣"
echo -e "║  Frontend  →  http://localhost:${FRONTEND_PORT}                  ║"
echo -e "║  Backend   →  http://localhost:${API_PORT_HOST}                 ║"
echo -e "║  Google Mock→  http://localhost:8090                ║"
echo -e "║  DB        →  localhost:5432  (postgres/postgres)   ║"
echo -e "╠══════════════════════════════════════════════════════╣"
echo -e "║  Admin login:                                        ║"
echo -e "║    Email:    ${ADMIN_EMAIL}         ║"
echo -e "║    Password: ${ADMIN_PASS}                          ║"
echo -e "╠══════════════════════════════════════════════════════╣"
echo -e "║  Commands:                                           ║"
echo -e "║    ./run_dev.sh stop   — stop everything             ║"
echo -e "║    ./run_dev.sh logs   — tail all logs               ║"
echo -e "╚══════════════════════════════════════════════════════╝${NC}"
echo ""
