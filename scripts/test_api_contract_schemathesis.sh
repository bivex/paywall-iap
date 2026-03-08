#!/usr/bin/env bash
# Contract-test API endpoints via Schemathesis against an OpenAPI schema.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
API_BASE="${API_BASE:-http://localhost:8081}"
API_BASE="${API_BASE%/}"
SCHEMA_URL="${SCHEMA_URL:-}"
SCHEMA_PATH="${SCHEMA_PATH:-}"
SCHEMATHESIS_AUTH_TOKEN="${SCHEMATHESIS_AUTH_TOKEN:-}"
SCHEMATHESIS_USER_AUTH_TOKEN="${SCHEMATHESIS_USER_AUTH_TOKEN:-}"
SCHEMATHESIS_HEADER="${SCHEMATHESIS_HEADER:-}"
EXPLICIT_SCHEMATHESIS_AUTH_TOKEN="$SCHEMATHESIS_AUTH_TOKEN"
EXPLICIT_SCHEMATHESIS_HEADER="$SCHEMATHESIS_HEADER"
SCHEMATHESIS_PHASES="${SCHEMATHESIS_PHASES:-examples,coverage,fuzzing}"
SKIP_HEALTHCHECK="${SKIP_HEALTHCHECK:-0}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@paywall.local}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin12345}"
APPLE_MOCK_BASE="${APPLE_MOCK_BASE:-http://localhost:9090}"
IAP_PRODUCT_ID="${IAP_PRODUCT_ID:-com.example.unlimited.1mo}"
DB_CONTAINER="${DB_CONTAINER:-paywall-db-1}"
DATABASE_URL="${DATABASE_URL:-}"
DB_NAME="${DB_NAME:-iap_db}"
DB_USER="${DB_USER:-postgres}"

if [[ -z "$SCHEMA_URL" && -z "$SCHEMA_PATH" ]]; then
  SCHEMA_URL="$API_BASE/openapi.yaml"
fi

say() { printf '\n==> %s\n' "$*"; }
fail() { echo "ERROR: $*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || fail "Missing required command: $1"
}

resolve_schema_target() {
  if [[ -n "$SCHEMA_URL" ]]; then
    printf '%s\n' "$SCHEMA_URL"
    return 0
  fi

  if [[ -n "$SCHEMA_PATH" ]]; then
    if [[ "$SCHEMA_PATH" = /* ]]; then
      printf '%s\n' "$SCHEMA_PATH"
    else
      printf '%s\n' "$ROOT_DIR/$SCHEMA_PATH"
    fi
    return 0
  fi

  local candidate
  for candidate in \
    "$ROOT_DIR/docs/api/openapi.yaml" \
    "$ROOT_DIR/docs/api/openapi.yml" \
    "$ROOT_DIR/backend/docs/openapi/openapi.yaml" \
    "$ROOT_DIR/backend/docs/openapi.yaml" \
    "$ROOT_DIR/backend/docs/openapi.yml" \
    "$ROOT_DIR/api/openapi.yaml" \
    "$ROOT_DIR/api/openapi.yml"
  do
    if [[ -f "$candidate" ]]; then
      printf '%s\n' "$candidate"
      return 0
    fi
  done

  return 1
}

detect_schemathesis_cmd() {
  if command -v schemathesis >/dev/null 2>&1; then
    SCHEMATHESIS_CMD=(schemathesis)
  elif command -v st >/dev/null 2>&1; then
    SCHEMATHESIS_CMD=(st)
  elif python3 -c 'import schemathesis' >/dev/null 2>&1; then
    SCHEMATHESIS_CMD=(python3 -m schemathesis)
  else
    fail "Schemathesis is not installed. Install it and rerun, or provide your own launcher via PATH."
  fi
}

detect_base_url_flag() {
  local help_output
  help_output="$("${SCHEMATHESIS_CMD[@]}" run --help 2>&1 || true)"

  if grep -q -- '--base-url' <<<"$help_output"; then
    SCHEMATHESIS_BASE_URL_FLAG="--base-url"
  elif grep -q -- '--url' <<<"$help_output"; then
    SCHEMATHESIS_BASE_URL_FLAG="--url"
  else
    fail "Could not detect the Schemathesis base URL flag from 'run --help'."
  fi
}

try_fetch_admin_token() {
  command -v curl >/dev/null 2>&1 || return 0
  command -v python3 >/dev/null 2>&1 || return 0

  local login_payload response token
  login_payload=$(printf '{"email":"%s","password":"%s"}' "$ADMIN_EMAIL" "$ADMIN_PASSWORD")
  response="$(curl -sS -X POST "$API_BASE/v1/admin/auth/login" -H 'Content-Type: application/json' -d "$login_payload" || true)"

  token="$(RESPONSE_JSON="$response" python3 - <<'PY'
import json, os
raw = os.environ.get('RESPONSE_JSON', '')
try:
    payload = json.loads(raw)
except Exception:
    print('')
    raise SystemExit(0)
data = payload.get('data') or {}
token = data.get('access_token') or payload.get('access_token') or ''
print(token)
PY
)"

  if [[ -n "$token" ]]; then
    SCHEMATHESIS_AUTH_TOKEN="$token"
    say "Using admin bearer token obtained via /v1/admin/auth/login"
  else
    say "Could not obtain admin bearer token automatically; running without explicit auth header"
  fi
}

try_fetch_user_token() {
  command -v curl >/dev/null 2>&1 || return 0
  command -v python3 >/dev/null 2>&1 || return 0

  local suffix register_payload response token
  suffix="$(python3 - <<'PY'
import uuid
print(uuid.uuid4())
PY
)"
  register_payload=$(printf '{"platform_user_id":"schemathesis-%s","device_id":"device-%s","platform":"ios","app_version":"1.0.0"}' "$suffix" "$suffix")
  response="$(curl -sS -X POST "$API_BASE/v1/auth/register" -H 'Content-Type: application/json' -d "$register_payload" || true)"

  token="$(RESPONSE_JSON="$response" python3 - <<'PY'
import json, os
raw = os.environ.get('RESPONSE_JSON', '')
try:
    payload = json.loads(raw)
except Exception:
    print('')
    raise SystemExit(0)
data = payload.get('data') or {}
token = data.get('access_token') or payload.get('access_token') or ''
print(token)
PY
)"

  if [[ -n "$token" ]]; then
    SCHEMATHESIS_USER_AUTH_TOKEN="$token"
    say "Using user bearer token obtained via /v1/auth/register"
  else
    say "Could not obtain user bearer token automatically; user-protected endpoints will run without explicit user auth"
  fi
}

seed_user_subscription_via_iap() {
  command -v curl >/dev/null 2>&1 || return 1
  command -v python3 >/dev/null 2>&1 || return 1
  [[ -n "$SCHEMATHESIS_USER_AUTH_TOKEN" ]] || return 1

  local mock_response receipt_token verify_body verify_status
  mock_response="$(curl -sS -X POST "$APPLE_MOCK_BASE/subs" -H 'Content-Type: application/json' -d "{\"productId\":\"$IAP_PRODUCT_ID\"}" || true)"
  receipt_token="$(RESPONSE_JSON="$mock_response" python3 - <<'PY'
import json, os
raw = os.environ.get('RESPONSE_JSON', '')
try:
    payload = json.loads(raw)
except Exception:
    print('')
    raise SystemExit(0)
print(payload.get('receiptToken', ''))
PY
)"

  if [[ -z "$receipt_token" ]]; then
    say "Could not create a valid Apple mock receipt token automatically"
    return 1
  fi

  verify_body=$(printf '{"platform":"ios","receipt_data":"%s","product_id":"%s","transaction_id":""}' "$receipt_token" "$IAP_PRODUCT_ID")
  verify_status="$(curl -sS -o /tmp/schemathesis-verify-iap.json -w '%{http_code}' -X POST "$API_BASE/v1/verify/iap" \
    -H "Authorization: Bearer $SCHEMATHESIS_USER_AUTH_TOKEN" \
    -H 'Content-Type: application/json' \
    -d "$verify_body" || true)"

  if [[ "$verify_status" == "200" || "$verify_status" == "409" ]]; then
    say "Seeded a valid user subscription via /v1/verify/iap"
    return 0
  fi

  say "Could not seed subscription via /v1/verify/iap (status=$verify_status)"
  return 1
}

run_sql_file() {
  local sql_file="$1"

  if [[ -n "$DATABASE_URL" ]]; then
    need psql
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$sql_file"
    return 0
  fi

  need docker
  docker exec -i "$DB_CONTAINER" psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$DB_NAME" < "$sql_file"
}

should_split_runs() {
  if [[ -n "$EXPLICIT_SCHEMATHESIS_AUTH_TOKEN" || -n "$EXPLICIT_SCHEMATHESIS_HEADER" ]]; then
    return 1
  fi

  local arg
  for arg in "$@"; do
    case "$arg" in
      --include-*|--exclude-*|--include-by|--exclude-by)
        return 1
        ;;
    esac
  done

  return 0
}

should_apply_admin_experiment_negative_rejection_workarounds() {
  local args_joined
  args_joined=" $* "

  if [[ "$args_joined" == *" --exclude-path "*lifecycle-audit* || "$args_joined" == *" --exclude-path-regex "*lifecycle-audit* ]]; then
    return 1
  fi

  if [[ "$args_joined" != *" --include-path "* && "$args_joined" != *" --include-path-regex "* && "$args_joined" != *" --exclude-path "* && "$args_joined" != *" --exclude-path-regex "* ]]; then
    return 0
  fi

  if [[ "$args_joined" == *"/v1/admin/experiments"* || "$args_joined" == *"lifecycle-audit"* ]]; then
    return 0
  fi

  return 1
}

should_reseed_admin_experiment_fixtures() {
  local args_joined
  args_joined=" $* "

  if [[ "$args_joined" != *" --include-path "* && "$args_joined" != *" --include-path-regex "* && "$args_joined" != *" --exclude-path "* && "$args_joined" != *" --exclude-path-regex "* ]]; then
    return 0
  fi

  if [[ "$args_joined" == *"/v1/admin/experiments"* || "$args_joined" == *"lifecycle-audit"* ]]; then
    return 0
  fi

  return 1
}

reseed_admin_experiment_contract_fixtures() {
  say "Re-seeding deterministic admin experiment fixtures"
  run_sql_file "$SCRIPT_DIR/seed_experiment_test_data.sql"
}

collect_non_path_filter_args() {
  FILTERED_SCHEMATHESIS_ARGS=()

  while (($# > 0)); do
    case "$1" in
      --include-path|--include-path-regex|--exclude-path|--exclude-path-regex)
        shift
        if (($# > 0)); then
          shift
        fi
        ;;
      *)
        FILTERED_SCHEMATHESIS_ARGS+=("$1")
        shift
        ;;
    esac
  done
}

run_schemathesis() {
  local label="$1"
  local bearer_token="$2"
  shift 2

  local -a cmd
  cmd=("${SCHEMATHESIS_CMD[@]}" run "$schema_target" "$SCHEMATHESIS_BASE_URL_FLAG" "$API_BASE")
  cmd+=(--phases "$SCHEMATHESIS_PHASES")

  if [[ -n "$bearer_token" ]]; then
    cmd+=(--header "Authorization: Bearer $bearer_token")
  fi

  if [[ -n "$SCHEMATHESIS_HEADER" ]]; then
    cmd+=(--header "$SCHEMATHESIS_HEADER")
  fi

  if (($# > 0)); then
    cmd+=("$@")
  fi

  say "Running Schemathesis ($label)"
  echo "  schema:   $schema_target"
  echo "  api_base: $API_BASE"
  "${cmd[@]}"
}

run_schemathesis_with_admin_experiment_negative_rejection_workarounds() {
  local label="$1"
  local bearer_token="$2"
  shift 2

  local lifecycle_audit_path='/v1/admin/experiments/{id}/lifecycle-audit'
  local update_experiment_path='/v1/admin/experiments/{id}'

  collect_non_path_filter_args "$@"

  say "Applying admin experiment Schemathesis workarounds (skip false-positive negative_data_rejection on schema-mutated valid IDs)"

  run_schemathesis "$label (excluding endpoint-specific negative_data_rejection false-positives)" "$bearer_token" \
    --exclude-path "$update_experiment_path" \
    --exclude-path "$lifecycle_audit_path" \
    "$@"

  run_schemathesis "$label (experiment update without negative_data_rejection)" "$bearer_token" \
    --include-path "$update_experiment_path" \
    --exclude-checks negative_data_rejection \
    "${FILTERED_SCHEMATHESIS_ARGS[@]}"

  run_schemathesis "$label (lifecycle-audit without negative_data_rejection)" "$bearer_token" \
    --include-path "$lifecycle_audit_path" \
    --exclude-checks negative_data_rejection \
    "${FILTERED_SCHEMATHESIS_ARGS[@]}"
}

main() {
  local schema_target

  detect_schemathesis_cmd
  detect_base_url_flag
  schema_target="$(resolve_schema_target)" || fail "OpenAPI schema not found. Set SCHEMA_URL or SCHEMA_PATH before running this script."

  if [[ "$schema_target" != http://* && "$schema_target" != https://* && ! -f "$schema_target" ]]; then
    fail "Schema target does not exist: $schema_target"
  fi

  if [[ "$SKIP_HEALTHCHECK" != "1" ]]; then
    command -v curl >/dev/null 2>&1 || fail "Missing required command: curl"
    say "Checking API health"
    curl -fsS "$API_BASE/health" >/dev/null || fail "Health check failed for $API_BASE/health"
  fi

  if [[ -z "$SCHEMATHESIS_AUTH_TOKEN" && -z "$SCHEMATHESIS_HEADER" ]]; then
    try_fetch_admin_token
  fi

  if should_split_runs "$@"; then
    try_fetch_user_token

    local user_suite_seeded=0
    if seed_user_subscription_via_iap; then
      user_suite_seeded=1
    fi

    run_schemathesis "public endpoints" "" \
      --exclude-tag admin \
      --exclude-tag admin-auth \
      --exclude-tag subscription \
      --exclude-tag iap \
      "$@"

    if should_reseed_admin_experiment_fixtures "$@"; then
      reseed_admin_experiment_contract_fixtures
    fi

    if should_apply_admin_experiment_negative_rejection_workarounds "$@"; then
      run_schemathesis_with_admin_experiment_negative_rejection_workarounds "admin endpoints" "$SCHEMATHESIS_AUTH_TOKEN" \
        --include-tag admin \
        "$@"
    else
      run_schemathesis "admin endpoints" "$SCHEMATHESIS_AUTH_TOKEN" \
        --include-tag admin \
        "$@"
    fi

    run_schemathesis "admin auth endpoints" "$SCHEMATHESIS_AUTH_TOKEN" \
      --include-tag admin-auth \
      "$@"

    if [[ "$user_suite_seeded" == "1" ]]; then
      run_schemathesis "user-protected endpoints" "$SCHEMATHESIS_USER_AUTH_TOKEN" \
        --include-tag subscription \
        "$@"
    else
      run_schemathesis "user-protected endpoints" "$SCHEMATHESIS_USER_AUTH_TOKEN" \
        --include-tag subscription \
        --include-tag iap \
        "$@"
    fi

    return 0
  fi

  if should_reseed_admin_experiment_fixtures "$@"; then
    reseed_admin_experiment_contract_fixtures
  fi

  if should_apply_admin_experiment_negative_rejection_workarounds "$@"; then
    run_schemathesis_with_admin_experiment_negative_rejection_workarounds "default" "$SCHEMATHESIS_AUTH_TOKEN" "$@"
  else
    run_schemathesis "default" "$SCHEMATHESIS_AUTH_TOKEN" "$@"
  fi
}

main "$@"