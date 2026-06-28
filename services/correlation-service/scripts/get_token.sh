#!/usr/bin/env sh
# Requires AuthService running. Set AUTH_SERVICE_URL, SERVICE_ID, SERVICE_SECRET in .env or compose.env.

set -e
. "${0%/*}/../.env" 2>/dev/null || . "${0%/*}/../compose.env" 2>/dev/null || true

AUTH_URL="${AUTH_SERVICE_URL:-http://localhost:8080/api/v1}"
SVC_ID="${SERVICE_ID:-correlation-service}"
SVC_SECRET="${SERVICE_SECRET:?Set SERVICE_SECRET in .env or compose.env}"
AUDIENCE="${JWT_AUDIENCE:-correlation-service}"

curl -sf -X POST "${AUTH_URL}/auth/login" \
  -H "X-Grant-Type: service" \
  -H "Content-Type: application/json" \
  -d "{\"service_id\":\"${SVC_ID}\",\"service_secret\":\"${SVC_SECRET}\",\"audience\":\"${AUDIENCE}\"}" \
  | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4
