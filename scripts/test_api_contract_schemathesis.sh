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
SCHEMATHESIS_HEADER="${SCHEMATHESIS_HEADER:-}"
SKIP_HEALTHCHECK="${SKIP_HEALTHCHECK:-0}"

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

main() {
  local schema_target
  local -a cmd

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

  say "Running Schemathesis"
  echo "  schema:   $schema_target"
  echo "  api_base: $API_BASE"

  cmd=("${SCHEMATHESIS_CMD[@]}" run "$schema_target" "$SCHEMATHESIS_BASE_URL_FLAG" "$API_BASE")

  if [[ -n "$SCHEMATHESIS_AUTH_TOKEN" ]]; then
    cmd+=(--header "Authorization: Bearer $SCHEMATHESIS_AUTH_TOKEN")
  fi

  if [[ -n "$SCHEMATHESIS_HEADER" ]]; then
    cmd+=(--header "$SCHEMATHESIS_HEADER")
  fi

  if (($# > 0)); then
    cmd+=("$@")
  fi

  "${cmd[@]}"
}

main "$@"