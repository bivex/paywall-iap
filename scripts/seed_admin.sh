#!/usr/bin/env bash
# scripts/seed_admin.sh — create or update first superadmin for admin dashboard
# Works with direct DB access OR via Docker container
#
# Usage (direct DB):
#   DATABASE_URL=postgres://... ./scripts/seed_admin.sh admin@example.com MyPassword123
#
# Usage (Docker container):
#   DB_CONTAINER=paywall-db-1 ./scripts/seed_admin.sh admin@example.com MyPassword123

set -euo pipefail

EMAIL="${1:-}"
PASSWORD="${2:-}"
DB_CONTAINER="${DB_CONTAINER:-paywall-db-1}"
DATABASE_URL="${DATABASE_URL:-}"

if [[ -z "$EMAIL" || -z "$PASSWORD" ]]; then
  echo "Usage: $0 <email> <password>"
  echo "   or: DATABASE_URL=postgres://... $0 <email> <password>"
  exit 1
fi

if [[ ${#PASSWORD} -lt 8 ]]; then
  echo "Error: password must be at least 8 characters"
  exit 1
fi

# Generate bcrypt hash via Go
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")/backend"

HASH=$(cd "$BACKEND_DIR" && cat > /tmp/genhash_seed.go << 'GOEOF'
package main
import ("fmt"; "os"; "golang.org/x/crypto/bcrypt")
func main() {
    h, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), 10)
    if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
    fmt.Println(string(h))
}
GOEOF
go run /tmp/genhash_seed.go "$PASSWORD")

echo "🔑 Generated bcrypt hash"

run_sql() {
  local sql="$1"
  if [[ -n "$DATABASE_URL" ]]; then
    psql "$DATABASE_URL" -c "$sql"
  else
    docker exec "$DB_CONTAINER" psql -U postgres -d iap_db -c "$sql"
  fi
}

run_sql_quiet() {
  local sql="$1"
  if [[ -n "$DATABASE_URL" ]]; then
    psql "$DATABASE_URL" -tAc "$sql"
  else
    docker exec "$DB_CONTAINER" psql -U postgres -d iap_db -tAc "$sql"
  fi
}

# Upsert user
echo "👤 Creating user: $EMAIL"
run_sql "
INSERT INTO users (platform_user_id, platform, app_version, email, role)
VALUES ('admin_web_$(echo "$EMAIL" | tr '@' '_')', 'web', '1.0.0', '$EMAIL', 'superadmin')
ON CONFLICT (email) DO UPDATE SET role = 'superadmin'
RETURNING id, email, role;
"

USER_ID=$(run_sql_quiet "SELECT id FROM users WHERE email='$EMAIL' LIMIT 1")

if [[ -z "$USER_ID" ]]; then
  echo "Error: failed to get user ID"
  exit 1
fi

echo "🔒 Setting password for user: $USER_ID"
run_sql "
INSERT INTO admin_credentials (user_id, password_hash)
VALUES ('$USER_ID', '$HASH')
ON CONFLICT (user_id) DO UPDATE
  SET password_hash = EXCLUDED.password_hash,
      updated_at    = now()
RETURNING user_id, created_at;
"

echo ""
echo "✅ Admin seeded successfully!"
echo "   Email:    $EMAIL"
echo "   Password: ${PASSWORD:0:2}$(printf '*%.0s' $(seq 1 $((${#PASSWORD}-2))))"
echo "   User ID:  $USER_ID"
echo "   Role:     superadmin"
echo ""
echo "👉 Login at: http://localhost:3000/auth/v1/login"
