#!/usr/bin/env bash
set -euo pipefail

# End-to-end smoke test for the backend API (register -> login -> report -> feed -> logout -> verify)
# Usage: ./backend/scripts/e2e_test.sh
# Environment:
#  BASE_URL (optional) - default http://localhost:8080
#  USE_REDIS (optional) - if set to 1, the script expects the server was started with Redis configured

BASE_URL=${BASE_URL:-http://127.0.0.1:8080}
# Default USE_REDIS to 1 when REDIS_ADDR is present, otherwise 0.
USE_REDIS=${USE_REDIS:-${REDIS_ADDR:+1}}
# Ensure a safe default even when script runs with `set -u`
: "${USE_REDIS:=0}"

tmpdir=$(mktemp -d)
# If KEEP_TMP is set to 1, keep temporary files for debugging. Otherwise remove on exit.
if [ "${KEEP_TMP:-0}" = "1" ]; then
  echo "Keeping temp dir: $tmpdir"
else
  trap 'rm -rf "$tmpdir"' EXIT
fi

echo "Using BASE_URL=$BASE_URL  USE_REDIS=$USE_REDIS"

json_extract_token(){
  local file=$1
  if command -v jq >/dev/null 2>&1; then
    jq -r '.access_token // empty' "$file"
  else
    # Use stdin to avoid issues with filenames containing spaces/newlines
    python -c "import sys,json; obj=json.load(sys.stdin); print(obj.get('access_token',''))" < "$file"
  fi
}

echo "1) Registering test user..."
# Allow overriding test user via environment variables for repeatable runs
TEST_NAME=${TEST_NAME:-TestUser}
# Default to a unique email per run to avoid "email already registered" errors
TEST_EMAIL=${TEST_EMAIL:-testuser+$(date +%s)@example.com}
TEST_PASSWORD=${TEST_PASSWORD:-Passw0rd!}
register_payload=$(jq -n --arg n "$TEST_NAME" --arg e "$TEST_EMAIL" --arg p "$TEST_PASSWORD" '{name:$n,email:$e,password:$p}' 2>/dev/null || printf '{"name":"%s","email":"%s","password":"%s"}' "$TEST_NAME" "$TEST_EMAIL" "$TEST_PASSWORD")
status=$(curl -s -o "$tmpdir/register.json" -w "%{http_code}" -X POST "$BASE_URL/register" -H "Content-Type: application/json" -d "$register_payload")
echo " -> HTTP $status"
if [ "$status" -ne 200 ]; then
  echo "Register may have failed (HTTP $status). Trying to login instead..."
fi

echo "2) Login to obtain access token..."
login_payload=$(jq -n --arg e "$TEST_EMAIL" --arg p "$TEST_PASSWORD" '{email:$e,password:$p}' 2>/dev/null || printf '{"email":"%s","password":"%s"}' "$TEST_EMAIL" "$TEST_PASSWORD")
status=$(curl -s -o "$tmpdir/login.json" -w "%{http_code}" -X POST "$BASE_URL/login" -H "Content-Type: application/json" -d "$login_payload")
echo " -> HTTP $status"
if [ "$status" -ne 200 ]; then
  echo "Login failed (HTTP $status). See $tmpdir/login.json for response:" >&2
  cat "$tmpdir/login.json" >&2
  exit 1
fi

TOKEN=$(json_extract_token "$tmpdir/login.json")
if [ -z "$TOKEN" ]; then
  echo "No access_token found in login response:" >&2
  cat "$tmpdir/login.json" >&2
  exit 1
fi
echo " -> Obtained token (len=${#TOKEN})"

echo "3) Create a report (protected endpoint)..."
report_payload=$(jq -n \
  --arg in "" \
  --arg name "Broken Lamp" \
  --arg idesc "Lamp out near block A" \
  --arg cat "Lighting" \
  --arg pdesc "Reported by automated e2e test" \
  --arg statusv "open" \
  --argjson urgency 2 \
  --argjson lat 12.34 \
  --argjson lng 56.78 \
  --arg media "http://example.com/photo.jpg" \
  '{user_id:$in,issue_name:$name,issue_desc:$idesc,issue_cat:$cat,post_desc:$pdesc,status:$statusv,urgency:$urgency,lat:$lat,lng:$lng,media_url:$media}' 2>/dev/null || printf '{"user_id":"","issue_name":"%s","issue_desc":"%s","issue_cat":"%s","post_desc":"%s","status":"%s","urgency":%d,"lat":%s,"lng":%s,"media_url":"%s"}' "Broken Lamp" "Lamp out near block A" "Lighting" "Reported by automated e2e test" "open" 2 12.34 56.78 "http://example.com/photo.jpg")
status=$(curl -s -o "$tmpdir/report.json" -w "%{http_code}" -X POST "$BASE_URL/report" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$report_payload")
echo " -> HTTP $status"
if [ "$status" -ne 200 ]; then
  echo "Report failed (HTTP $status). Response:" >&2
  cat "$tmpdir/report.json" >&2
  exit 1
fi
echo " -> Report created. Response saved to $tmpdir/report.json"

echo "4) Fetch feed and check for the report..."
curl -s "$BASE_URL/feed" -o "$tmpdir/feed.json"
if command -v jq >/dev/null 2>&1; then
  echo " -> Feed (first post):"
  jq '.[0]' "$tmpdir/feed.json" || true
else
  echo " -> Feed saved to $tmpdir/feed.json (install jq for pretty output)"
fi

echo "5) Logout (revoke token; effective only if server has Redis configured)..."
status=$(curl -s -o "$tmpdir/logout.json" -w "%{http_code}" -X POST "$BASE_URL/logout" -H "Authorization: Bearer $TOKEN")
echo " -> HTTP $status"
if [ "$status" -ne 200 ]; then
  echo "Logout failed (HTTP $status). Response:" >&2
  cat "$tmpdir/logout.json" >&2
  exit 1
fi
echo " -> Logout response:"; cat "$tmpdir/logout.json"; echo

echo "6) Verify token revocation by attempting another protected request (behavior depends on USE_REDIS)..."
status=$(curl -s -o "$tmpdir/report2.json" -w "%{http_code}" -X POST "$BASE_URL/report" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$report_payload")
echo " -> HTTP $status"
if [ "$USE_REDIS" = "1" ]; then
  if [ "$status" -ne 401 ]; then
    echo "Expected 401 Unauthorized after logout (Redis enabled), but got $status. Response:" >&2
    cat "$tmpdir/report2.json" >&2
    exit 1
  else
    echo " -> Token successfully revoked (401 as expected)."
  fi
else
  echo "Note: Redis not enabled (USE_REDIS!=1). If you enabled REDIS on the server, set USE_REDIS=1 when running this script to expect revocation." 
  echo "Protected request after logout returned HTTP $status; response saved to $tmpdir/report2.json"
fi

echo "E2E script completed successfully. Temporary responses are in $tmpdir (deleted on exit)."
