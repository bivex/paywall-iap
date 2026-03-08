#!/usr/bin/env bash
# run_dev.sh — cold-start the full paywall-iap dev stack
#
# What it does:
#   1. Ensures Docker Desktop is running
#   2. Tears down any stale containers (orphans too)
#   3. Releases any processes holding the required ports
#   4. Starts postgres + redis + mocks + api + worker
#   5. Runs migrations, auto-fixes known dirty-flag issues
#   6. Seeds full cold-start local test data
#   7. Starts Next.js frontend with hot-reload
#   8. Prints health summary + URLs
#
# Usage:
#   ./run_dev.sh                         # defaults: admin@paywall.local / admin12345
#   ./run_dev.sh stop                    # stop everything
#   ./run_dev.sh logs                    # tail all logs
#   API_PORT_HOST=8082 ./run_dev.sh      # override API host port
#   ADMIN_EMAIL=me@x.com ADMIN_PASS=secret ./run_dev.sh

set -euo pipefail

# ─── Config ──────────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_COMPOSE="$SCRIPT_DIR/infra/docker-compose/docker-compose.local.yml"
FRONTEND_COMPOSE="$SCRIPT_DIR/frontend/docker-compose.dev.yml"
DB_CONTAINER="paywall-db-1"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@paywall.local}"
ADMIN_PASS="${ADMIN_PASS:-admin12345}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"

# Fixed ports for mocks (no dynamic selection needed)
GOOGLE_MOCK_PORT=8090
APPLE_MOCK_PORT=9090

# API: default 8081, but user can override
API_PORT_HOST="${API_PORT_HOST:-8081}"
API_PORT_INTERNAL=8080

# ─── Colors ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
info()    { echo -e "${CYAN}▶ $*${NC}"; }
ok()      { echo -e "${GREEN}✔ $*${NC}"; }
warn()    { echo -e "${YELLOW}⚠ $*${NC}"; }
die()     { echo -e "${RED}✖ $*${NC}" >&2; exit 1; }

# ─── Helpers ─────────────────────────────────────────────────────────────────

# Kill any non-Docker process holding a port (Docker containers are handled by compose down)
release_port() {
  local port="$1"
  local pids
  pids=$(lsof -ti "TCP:${port}" -sTCP:LISTEN 2>/dev/null || true)
  if [[ -n "$pids" ]]; then
    warn "Port $port held by PIDs $pids — killing..."
    echo "$pids" | xargs kill -9 2>/dev/null || true
    sleep 0.5
  fi
}

# Stop any Docker container whose port binding matches a given host port
release_docker_port() {
  local port="$1"
  local cids
  cids=$(docker ps -q --filter "publish=$port" 2>/dev/null || true)
  if [[ -n "$cids" ]]; then
    warn "Docker containers using port $port — stopping..."
    echo "$cids" | xargs docker stop 2>/dev/null || true
    sleep 1
  fi
}

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
  docker compose -f "$BACKEND_COMPOSE"  down --remove-orphans -t 5 2>/dev/null || true
  docker compose -f "$FRONTEND_COMPOSE" down --remove-orphans -t 5 2>/dev/null || true
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
info "Stopping existing containers..."
docker compose -f "$BACKEND_COMPOSE"  down --remove-orphans -t 5 2>/dev/null || true
docker compose -f "$FRONTEND_COMPOSE" down --remove-orphans -t 5 2>/dev/null || true
ok "Containers stopped"

# ── Step 3: Release ports ─────────────────────────────────────────────────────
info "Releasing ports $API_PORT_HOST, $GOOGLE_MOCK_PORT, $APPLE_MOCK_PORT, $FRONTEND_PORT..."
for port in "$API_PORT_HOST" "$GOOGLE_MOCK_PORT" "$APPLE_MOCK_PORT" "$FRONTEND_PORT" 5432 6379; do
  release_docker_port "$port"
  release_port "$port"
done
ok "Ports clear"

# ── Step 4: Start backend stack ───────────────────────────────────────────────
info "Starting backend stack (API port: $API_PORT_HOST)..."
export API_PORT_HOST
docker compose -f "$BACKEND_COMPOSE" up -d --build

wait_for_healthy "$DB_CONTAINER" 60
ok "Backend stack up"

# ── Step 5: Run migrations ────────────────────────────────────────────────────
info "Running migrations..."

# Fix: remove now() from partial index predicates (not IMMUTABLE in Postgres)
sed -i.bak \
  "s/WHERE status = 'pending' AND next_retry_at <= now()/WHERE status = 'pending'/g;
   s/WHERE status = 'sent' AND sent_at < NOW() - INTERVAL '30 days'/WHERE status = 'sent'/g" \
  "$SCRIPT_DIR/backend/migrations/016_create_matomo_staged_events.up.sql" 2>/dev/null || true
rm -f "$SCRIPT_DIR/backend/migrations/016_create_matomo_staged_events.up.sql.bak"

# Fix: partial UNIQUE inside CREATE TABLE is unsupported
sed -i.bak \
  "s/CONSTRAINT unique_active_assignment UNIQUE (experiment_id, user_id)[[:space:]]*WHERE expires_at > now()/CONSTRAINT unique_assignment UNIQUE (experiment_id, user_id)/g" \
  "$SCRIPT_DIR/backend/migrations/015_create_ab_test_assignments.up.sql" 2>/dev/null || true
rm -f "$SCRIPT_DIR/backend/migrations/015_create_ab_test_assignments.up.sql.bak"

# Clear dirty flag before migrator run
docker exec "$DB_CONTAINER" psql -U postgres -d iap_db \
  -c "UPDATE schema_migrations SET dirty = false WHERE dirty = true;" &>/dev/null || true

# Recover known partial migration gaps from earlier broken runs.
# If migration 015 was skipped after a dirty-state reset, the table can be missing
# even though schema_migrations already advanced past version 15.
HAS_ASSIGNMENTS=$(docker exec "$DB_CONTAINER" psql -U postgres -d iap_db \
  -tAc "SELECT to_regclass('public.ab_test_assignments') IS NOT NULL;" 2>/dev/null || echo "f")
HAS_ASSIGNMENTS="${HAS_ASSIGNMENTS//[[:space:]]/}"
if [[ "$HAS_ASSIGNMENTS" != "t" ]]; then
  warn "ab_test_assignments is missing — applying migration 015 directly"
  docker exec -i "$DB_CONTAINER" psql -v ON_ERROR_STOP=1 -U postgres -d iap_db \
    < "$SCRIPT_DIR/backend/migrations/015_create_ab_test_assignments.up.sql"
fi

info "Building migrator image..."
docker compose -f "$BACKEND_COMPOSE" build migrator 2>&1 | tail -3
info "Applying migrations..."
if ! docker compose -f "$BACKEND_COMPOSE" run --rm migrator 2>&1 | tail -5; then
  warn "Migrator exited non-zero (may be already up-to-date) — continuing"
fi

HAS_WEB_PLATFORM=$(docker exec "$DB_CONTAINER" psql -U postgres -d iap_db \
  -tAc "SELECT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_platform_check' AND pg_get_constraintdef(oid) LIKE '%web%');" 2>/dev/null || echo "f")
HAS_WEB_PLATFORM="${HAS_WEB_PLATFORM//[[:space:]]/}"
if [[ "$HAS_WEB_PLATFORM" != "t" ]]; then
  warn "users_platform_check still does not allow web — applying migration 019 directly"
  docker exec -i "$DB_CONTAINER" psql -v ON_ERROR_STOP=1 -U postgres -d iap_db \
    < "$SCRIPT_DIR/backend/migrations/019_add_web_platform.up.sql"
fi

HAS_ASSIGNMENTS=$(docker exec "$DB_CONTAINER" psql -U postgres -d iap_db \
  -tAc "SELECT to_regclass('public.ab_test_assignments') IS NOT NULL;" 2>/dev/null || echo "f")
HAS_ASSIGNMENTS="${HAS_ASSIGNMENTS//[[:space:]]/}"
[[ "$HAS_ASSIGNMENTS" == "t" ]] || die "Migration recovery incomplete: ab_test_assignments is still missing"

HAS_WEB_PLATFORM=$(docker exec "$DB_CONTAINER" psql -U postgres -d iap_db \
  -tAc "SELECT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_platform_check' AND pg_get_constraintdef(oid) LIKE '%web%');" 2>/dev/null || echo "f")
HAS_WEB_PLATFORM="${HAS_WEB_PLATFORM//[[:space:]]/}"
[[ "$HAS_WEB_PLATFORM" == "t" ]] || die "Migration recovery incomplete: users platform constraint still rejects web"

HAS_WINDOW_EVENT_TABLE=$(docker exec "$DB_CONTAINER" psql -U postgres -d iap_db \
  -tAc "SELECT to_regclass('public.bandit_window_events') IS NOT NULL;" 2>/dev/null || echo "f")
HAS_WINDOW_EVENT_TABLE="${HAS_WINDOW_EVENT_TABLE//[[:space:]]/}"
if [[ "$HAS_WINDOW_EVENT_TABLE" == "t" ]]; then
  WINDOW_EVENT_TYPE_LEN=$(docker exec "$DB_CONTAINER" psql -U postgres -d iap_db \
    -tAc "SELECT COALESCE(character_maximum_length, 0) FROM information_schema.columns WHERE table_schema='public' AND table_name='bandit_window_events' AND column_name='event_type';" 2>/dev/null || echo "0")
  WINDOW_EVENT_TYPE_LEN="${WINDOW_EVENT_TYPE_LEN//[[:space:]]/}"
  if [[ "$WINDOW_EVENT_TYPE_LEN" -lt 13 ]]; then
    warn "bandit_window_events.event_type is too narrow (${WINDOW_EVENT_TYPE_LEN}) — widening to VARCHAR(20)"
    docker exec "$DB_CONTAINER" psql -v ON_ERROR_STOP=1 -U postgres -d iap_db \
      -c "ALTER TABLE bandit_window_events ALTER COLUMN event_type TYPE VARCHAR(20);"
  fi
fi

ok "Migrations done"

# ── Step 6: Seed cold-start local test data ────────────────────────────────────
info "Seeding cold-start local test data..."
DB_CONTAINER="$DB_CONTAINER" bash "$SCRIPT_DIR/scripts/seed_all_test_data.sh" "$ADMIN_EMAIL" "$ADMIN_PASS"
ok "Cold-start test data seeded"

# ── Step 7: Start frontend ─────────────────────────────────────────────────────
export FRONTEND_PORT
export NEXT_PUBLIC_API_URL="http://localhost:${API_PORT_HOST}"
export BACKEND_URL="http://paywall-api-1:8080"

info "Building frontend image (Next.js hot-reload)..."
docker compose -f "$FRONTEND_COMPOSE" build 2>&1 | tail -5

info "Starting frontend container..."
docker compose -f "$FRONTEND_COMPOSE" up -d
ok "Frontend container started"

# ── Step 8: Health checks ──────────────────────────────────────────────────────
wait_for_http "http://localhost:${API_PORT_HOST}/health" "Backend API" 30
wait_for_http "http://localhost:${FRONTEND_PORT}"        "Frontend"    90

# ── Done ───────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════╗"
echo -e "║  ✅  Dev stack is ready!                             ║"
echo -e "╠══════════════════════════════════════════════════════╣"
printf "${GREEN}║  Frontend    →  http://localhost:%-20s║\n" "${FRONTEND_PORT}"
printf "${GREEN}║  Backend API →  http://localhost:%-20s║\n" "${API_PORT_HOST}"
printf "${GREEN}║  Google Mock →  http://localhost:%-20s║\n" "${GOOGLE_MOCK_PORT}"
printf "${GREEN}║  Apple Mock  →  http://localhost:%-20s║\n" "${APPLE_MOCK_PORT}"
echo -e "║  DB          →  localhost:5432  (postgres/postgres) ║"
echo -e "╠══════════════════════════════════════════════════════╣"
printf "${GREEN}║  Admin email:    %-35s║\n" "$ADMIN_EMAIL"
printf "${GREEN}║  Admin password: %-35s║\n" "$ADMIN_PASS"
echo -e "╠══════════════════════════════════════════════════════╣"
echo -e "║  ./run_dev.sh stop  — stop everything               ║"
echo -e "║  ./run_dev.sh logs  — tail all logs                 ║"
echo -e "╚══════════════════════════════════════════════════════╝${NC}"
echo ""

