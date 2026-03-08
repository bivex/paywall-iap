#!/usr/bin/env bash
# scripts/seed_all_test_data.sh — one-shot cold-start local test data bootstrap
# Seeds admin credentials, pricing tiers, realistic dashboard fixtures,
# and deterministic experiment/bandit data for non-empty admin pages.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_CONTAINER="${DB_CONTAINER:-paywall-db-1}"
DATABASE_URL="${DATABASE_URL:-}"
DB_NAME="${DB_NAME:-iap_db}"
DB_USER="${DB_USER:-postgres}"
ADMIN_EMAIL="${1:-${ADMIN_EMAIL:-admin@paywall.local}}"
ADMIN_PASSWORD="${2:-${ADMIN_PASSWORD:-admin12345}}"

need() {
  command -v "$1" >/dev/null || {
    echo "Error: required command not found: $1" >&2
    exit 1
  }
}

run_sql_file() {
  local sql_file="$1"
  if [[ -n "$DATABASE_URL" ]]; then
    need psql
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$sql_file"
  else
    need docker
    docker exec -i "$DB_CONTAINER" psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$DB_NAME" < "$sql_file"
  fi
}

echo "🌱 Seeding full cold-start test data"
if [[ -n "$DATABASE_URL" ]]; then
  echo "   Mode: direct DATABASE_URL"
else
  echo "   Mode: docker container ($DB_CONTAINER)"
fi

echo ""
echo "1/4 → Admin credentials"
DB_CONTAINER="$DB_CONTAINER" DATABASE_URL="$DATABASE_URL" bash "$SCRIPT_DIR/seed_admin.sh" "$ADMIN_EMAIL" "$ADMIN_PASSWORD"

echo ""
echo "2/4 → Pricing tiers"
DB_CONTAINER="$DB_CONTAINER" DATABASE_URL="$DATABASE_URL" DB_NAME="$DB_NAME" DB_USER="$DB_USER" bash "$SCRIPT_DIR/seed_tiers.sh"

echo ""
echo "3/4 → Revenue/dashboard fixtures"
run_sql_file "$SCRIPT_DIR/seed_dev_data.sql"

echo ""
echo "4/4 → Experiment/bandit fixtures"
run_sql_file "$SCRIPT_DIR/seed_experiment_test_data.sql"

echo ""
echo "✅ Cold-start test data seeded successfully"
echo "   Admin login: $ADMIN_EMAIL / $ADMIN_PASSWORD"
echo "   Seed experiment IDs:"
echo "     - 10000000-0000-0000-0000-000000000001 (running hybrid)"
echo "     - 10000000-0000-0000-0000-000000000002 (draft classic)"
echo "     - 10000000-0000-0000-0000-000000000003 (paused bandit)"
echo ""
echo "👉 Suggested checks:"
echo "   - http://localhost:3000/dashboard/pricing"
echo "   - http://localhost:3000/dashboard/experiments"
echo "   - http://localhost:3000/dashboard/experiments/bandit"
echo "   - http://localhost:3000/dashboard/experiments/studio"
echo "   - http://localhost:3000/dashboard/experiments/feedback"
echo "   - http://localhost:3000/dashboard/experiments/sliding-window"
echo "   - http://localhost:3000/dashboard/experiments/multi-objective"