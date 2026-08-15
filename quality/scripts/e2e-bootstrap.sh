#!/usr/bin/env bash
# Bootstrap E2E: start demo compose stack, run demo.sh, tear down.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${ROOT}"

COMPOSE_FILE="docker-compose.demo.yml"
GATEWAY_PORT="${GATEWAY_PORT:-8080}"
export GATEWAY_URL="${GATEWAY_URL:-http://localhost:${GATEWAY_PORT}}"

cleanup() {
  docker compose -f "${COMPOSE_FILE}" down -v --remove-orphans >/dev/null 2>&1 || true
}

trap cleanup EXIT

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

require_cmd docker
require_cmd curl
require_cmd go

echo "[e2e-bootstrap] building and starting demo stack"
docker compose -f "${COMPOSE_FILE}" up --build -d

echo "[e2e-bootstrap] running demo scenarios"
bash "${ROOT}/scripts/demo.sh"

echo "[e2e-bootstrap] PASSED"
