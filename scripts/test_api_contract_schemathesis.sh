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

if [[ -z "$SCHEMA_URL" && -z "$SCHEMA_PATH" ]]; then
  SCHEMA_URL="$API_BASE/openapi.yaml"
fi

say() { printf '\n==> %s\n' "$*"; }
fail() { echo "ERROR: $*" >&2; exit 1; }

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

    run_schemathesis "public endpoints" "" \
      --exclude-tag admin \
      --exclude-tag admin-auth \
      --exclude-tag subscription \
      --exclude-tag iap \
      "$@"

    run_schemathesis "admin endpoints" "$SCHEMATHESIS_AUTH_TOKEN" \
      --include-tag admin \
      --include-tag admin-auth \
      "$@"

    run_schemathesis "user-protected endpoints" "$SCHEMATHESIS_USER_AUTH_TOKEN" \
      --include-tag subscription \
      --include-tag iap \
      "$@"

    return 0
  fi

  run_schemathesis "default" "$SCHEMATHESIS_AUTH_TOKEN" "$@"
}

main "$@"