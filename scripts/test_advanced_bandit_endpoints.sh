#!/usr/bin/env bash
# Smoke-test advanced bandit endpoints on the local dev stack.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
BACKEND_COMPOSE="${BACKEND_COMPOSE:-$ROOT_DIR/infra/docker-compose/docker-compose.local.yml}"
API_BASE="${API_BASE:-http://localhost:8081}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@paywall.local}"
ADMIN_PASS="${ADMIN_PASS:-admin12345}"
DB_NAME="${DB_NAME:-iap_db}"
TEST_USER_ID="${TEST_USER_ID:-22222222-2222-2222-2222-222222222222}"
RUN_SCHEMATHESIS="${RUN_SCHEMATHESIS:-0}"
SCHEMA_URL="${SCHEMA_URL:-}"
SCHEMA_PATH="${SCHEMA_PATH:-}"

BODY_FILE="$(mktemp)"
cleanup() { rm -f "$BODY_FILE"; }
trap cleanup EXIT

need() { command -v "$1" >/dev/null || { echo "Missing required command: $1" >&2; exit 1; }; }
need curl; need python3; need docker

say() { printf '\n==> %s\n' "$*"; }
fail() { echo "ERROR: $*" >&2; exit 1; }

json_get() {
  local path="$1"
  python3 -c '
import json, sys
obj = json.load(sys.stdin)
for part in sys.argv[1].split("."):
    obj = obj[part]
print(json.dumps(obj) if isinstance(obj, (dict, list)) else obj)
' "$path"
}

request() {
  local method="$1" url="$2" body="${3:-}" token="${4:-}"
  local cmd=(curl -sS -o "$BODY_FILE" -w '%{http_code}' -X "$method" "$url")
  [[ -n "$token" ]] && cmd+=(-H "Authorization: Bearer $token")
  if [[ -n "$body" ]]; then
    cmd+=(-H 'Content-Type: application/json' -d "$body")
  fi
  RESPONSE_STATUS="$("${cmd[@]}")"
  RESPONSE_BODY="$(cat "$BODY_FILE")"
}

assert_status() {
  local expected="$1"
  [[ "$RESPONSE_STATUS" == "$expected" ]] || fail "Expected HTTP $expected, got $RESPONSE_STATUS: $RESPONSE_BODY"
}

db_query() {
  docker compose -f "$BACKEND_COMPOSE" exec -T db psql -U postgres -d "$DB_NAME" -tAc "$1"
}

new_uuid() {
  python3 - <<'PY'
import uuid
print(uuid.uuid4())
PY
}

say "Checking API health"
request GET "$API_BASE/health"
assert_status 200

say "Logging in as admin"
request POST "$API_BASE/v1/admin/auth/login" "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}"
assert_status 200
TOKEN="$(printf '%s' "$RESPONSE_BODY" | json_get 'data.access_token')"
[[ -n "$TOKEN" ]] || fail "Admin login returned empty token"

EXPERIMENT_NAME="Advanced Bandit Smoke $(date +%Y%m%d-%H%M%S)"
CREATE_PAYLOAD=$(cat <<JSON
{"name":"$EXPERIMENT_NAME","description":"Local dev smoke test for advanced bandit endpoints","status":"running","algorithm_type":"thompson_sampling","is_bandit":true,"min_sample_size":200,"confidence_threshold_percent":95,"arms":[{"name":"Control","description":"Baseline","is_control":true,"traffic_weight":1},{"name":"Variant A","description":"Alternative","is_control":false,"traffic_weight":1}]}
JSON
)

say "Creating test experiment"
request POST "$API_BASE/v1/admin/experiments" "$CREATE_PAYLOAD" "$TOKEN"
assert_status 201
EXPERIMENT_ID="$(printf '%s' "$RESPONSE_BODY" | json_get 'data.id')"
[[ -n "$EXPERIMENT_ID" ]] || fail "Failed to parse experiment id"
ARM_ID="$(db_query "select id from ab_test_arms where experiment_id = '$EXPERIMENT_ID' order by is_control desc, created_at asc limit 1;")"
[[ -n "$ARM_ID" ]] || fail "Failed to load arm id from DB"

say "Reading initial objective scores"
request GET "$API_BASE/v1/bandit/experiments/$EXPERIMENT_ID/objectives"
assert_status 200

say "Saving hybrid objective config"
request PUT "$API_BASE/v1/bandit/experiments/$EXPERIMENT_ID/objectives/config" '{"objective_type":"hybrid","objective_weights":{"conversion":0.5,"ltv":0.3,"revenue":0.2}}'
assert_status 200

OBJECTIVE_TYPE="$(db_query "select objective_type from ab_tests where id = '$EXPERIMENT_ID';")"
[[ "$OBJECTIVE_TYPE" == "hybrid" ]] || fail "Expected DB objective_type=hybrid, got: $OBJECTIVE_TYPE"

say "Re-reading objective scores after config update"
request GET "$API_BASE/v1/bandit/experiments/$EXPERIMENT_ID/objectives"
assert_status 200
printf '%s' "$RESPONSE_BODY" | python3 -c '
import json, sys
payload = json.load(sys.stdin)
assert payload, "empty objective response"
for _, scores in payload.items():
    for key in ("conversion", "ltv", "revenue", "hybrid"):
        assert key in scores, f"missing {key} score"
'

for endpoint in \
  "$API_BASE/v1/bandit/experiments/$EXPERIMENT_ID/window/info" \
  "$API_BASE/v1/bandit/experiments/$EXPERIMENT_ID/window/events?limit=10" \
  "$API_BASE/v1/bandit/experiments/$EXPERIMENT_ID/metrics"
do
  say "GET $endpoint"
  request GET "$endpoint"
  assert_status 200
done

say "Trimming window"
request POST "$API_BASE/v1/bandit/experiments/$EXPERIMENT_ID/window/trim"
assert_status 200

PENDING_ID="$(new_uuid)"
say "Seeding one pending reward in DB"
db_query "insert into bandit_pending_rewards (id, experiment_id, arm_id, user_id, assigned_at, expires_at, converted) values ('$PENDING_ID','$EXPERIMENT_ID','$ARM_ID','$TEST_USER_ID', now(), now() + interval '7 days', false);"

say "Reading pending reward by id"
request GET "$API_BASE/v1/bandit/pending/$PENDING_ID"
assert_status 200

say "Reading pending rewards by user"
request GET "$API_BASE/v1/bandit/users/$TEST_USER_ID/pending"
assert_status 200

if [[ "$RUN_SCHEMATHESIS" == "1" ]]; then
  say "Running Schemathesis contract checks"
  API_BASE="$API_BASE" SCHEMA_URL="$SCHEMA_URL" SCHEMA_PATH="$SCHEMA_PATH" \
    bash "$SCRIPT_DIR/test_api_contract_schemathesis.sh"
fi

echo
echo "PASS"
echo "  experiment_id: $EXPERIMENT_ID"
echo "  arm_id:        $ARM_ID"
echo "  pending_id:    $PENDING_ID"
echo "  api_base:      $API_BASE"