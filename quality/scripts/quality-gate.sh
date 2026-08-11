#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

echo "=== Universal API Gateway Engineering Quality Gate ==="

echo "[1] Formatting"
UNFORMATTED="$(gofmt -l .)"
if [[ -n "$UNFORMATTED" ]]; then
  echo "FAIL: unformatted files:"
  echo "$UNFORMATTED"
  exit 1
fi
echo "PASS: formatting"

echo "[2] Static analysis"
go vet ./...
echo "PASS: static analysis"

echo "[3] Unit tests"
go test -race -count=1 ./...
echo "PASS: unit tests"

echo "[4] Coverage"
bash quality/scripts/check-coverage.sh

echo "[5] Mutation testing"
echo "TODO: select tool via ADR (Phase 3, see quality/config/mutation.yaml)"

echo "[6] Integration tests"
echo "TODO: Phase 2"

echo "[7] Regression tests"
echo "TODO: Phase 2 — every bug fix must add a permanent regression test"

echo "[8] E2E tests"
echo "TODO: Phase 2"

echo "[9] Dependencies"
bash quality/scripts/check-dependencies.sh

echo "[10] Security"
echo "TODO: Phase 4 — SAST, secret scanning, container scanning, SBOM"

echo "[11] Complexity and module size"
bash quality/scripts/check-complexity.sh
bash quality/scripts/check-module-size.sh

echo "[12] API compatibility and performance"
bash quality/scripts/check-api-compatibility.sh
bash quality/scripts/check-performance.sh

bash quality/scripts/generate-report.sh

echo "QUALITY GATE: PASSED (Phase 1)"
