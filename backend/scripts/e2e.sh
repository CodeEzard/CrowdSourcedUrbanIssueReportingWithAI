#!/usr/bin/env bash
set -euo pipefail

# E2E helper: idempotent register/login -> post report -> fetch feed
# Usage: ./backend/scripts/e2e.sh [base_url]

BASE_URL=${1:-http://localhost:8080}
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

REG_FILE="$TMPDIR/register.json"
LOGIN_FILE="$TMPDIR/login.json"
REPORT_FILE="$TMPDIR/report.json"

cat > "$REG_FILE" <<'JSON'
{"name":"E2EUser","email":"e2e+test@example.com","password":"password"}
JSON

cat > "$LOGIN_FILE" <<'JSON'
{"email":"e2e+test@example.com","password":"password"}
JSON

cat > "$REPORT_FILE" <<'JSON'
{
  "issue_name": "Automation Test Issue",
  "issue_desc": "Created by automated test",
  "issue_cat": "Road",
  "post_desc": "Initial report from test",
  "status": "open",
  "urgency": 2,
  "lat": 12.34,
  "lng": 56.78,
  "media_url": "http://example.com/test.jpg"
}
JSON

echo "Using base URL: $BASE_URL"

http() { # simple wrapper to call curl and print action
  curl -sS -w "\nHTTP_STATUS:%{http_code}\n" "$@"
}

extract_token_from_stdin() {
  # try jq, then python, then grep
  if command -v jq >/dev/null 2>&1; then
    jq -r .access_token 2>/dev/null || true
    return
  fi
  python - <<'PY' 2>/dev/null || true
import sys, json
try:
    obj=json.load(sys.stdin)
    print(obj.get('access_token',''))
except Exception:
    pass
PY
  # fallback grep
  grep -oP '"access_token"\s*:\s*"\K[^"]+' || true
}

echo "== REGISTER (idempotent) =="
REG_RESP=$(http -X POST -H "Content-Type: application/json" --data-binary @"$REG_FILE" "$BASE_URL/register" ) || true
echo "$REG_RESP"
REG_STATUS=$(echo "$REG_RESP" | sed -n 's/.*HTTP_STATUS:\([0-9][0-9][0-9]\)/\1/p')

if [ "$REG_STATUS" = "200" ]; then
  echo "Registered successfully"
  TOKEN=$(echo "$REG_RESP" | sed -n '1,/HTTP_STATUS:/p' | extract_token_from_stdin)
fi

if [ -z "${TOKEN-}" ]; then
  echo "Register did not return token (status=$REG_STATUS), trying login..."
  LOGIN_RESP=$(http -X POST -H "Content-Type: application/json" --data-binary @"$LOGIN_FILE" "$BASE_URL/login" ) || true
  echo "$LOGIN_RESP"
  LOGIN_STATUS=$(echo "$LOGIN_RESP" | sed -n 's/.*HTTP_STATUS:\([0-9][0-9][0-9]\)/\1/p')
  if [ "$LOGIN_STATUS" != "200" ]; then
    echo "Login failed with status $LOGIN_STATUS" >&2
    exit 1
  fi
  TOKEN=$(echo "$LOGIN_RESP" | sed -n '1,/HTTP_STATUS:/p' | extract_token_from_stdin)
fi

if [ -z "$TOKEN" ]; then
  echo "Failed to obtain access token" >&2
  exit 1
fi

echo "TOKEN=$TOKEN"

echo "== POST REPORT =="
REPORT_RESP=$(http -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" --data-binary @"$REPORT_FILE" "$BASE_URL/report" ) || true
echo "$REPORT_RESP"
REPORT_STATUS=$(echo "$REPORT_RESP" | sed -n 's/.*HTTP_STATUS:\([0-9][0-9][0-9]\)/\1/p')
echo "Report HTTP status: $REPORT_STATUS"

echo "== FETCH FEED =="
FEED=$(curl -sS -H "Accept: application/json" "$BASE_URL/feed" || true)
if [ -n "$FEED" ]; then
  if command -v jq >/dev/null 2>&1; then
    echo "$FEED" | jq '.'
  else
    echo "$FEED" | python -m json.tool || echo "$FEED"
  fi
else
  echo "(no feed response)"
fi

echo "E2E script finished"
