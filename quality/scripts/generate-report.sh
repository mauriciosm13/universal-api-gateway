#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

mkdir -p quality/reports

GLOBAL=""
if [[ -f coverage.out ]]; then
  GLOBAL="$(go tool cover -func=coverage.out | awk '/^total:/ {gsub(/%/,"",$3); print $3}')"
fi

GLOBAL_MIN="$(awk '/global_minimum:/ {print $2}' quality/config/coverage.yaml)"
CHANGED_MIN="$(awk '/changed_code_minimum:/ {print $2}' quality/config/coverage.yaml)"

cat > quality/reports/latest.md <<EOF
# Quality Report

Status: PHASE1

| Gate | Status |
|---|---|
| Formatting | PASS (when gate completes) |
| Static analysis | PASS (when gate completes) |
| Unit tests | PASS (when gate completes) |
| Global coverage | ${GLOBAL:-n/a}% / ${GLOBAL_MIN}% minimum |
| Changed packages | ${CHANGED_MIN}% soft minimum on PRs |
| Mutation | TODO Phase 3 |
| Integration / E2E | TODO Phase 2 |
| Security | TODO Phase 4 |
| Complexity / size | TODO Phase 3 |
| API / performance | TODO Phase 5 |

Full machine-readable JSON reports arrive in Phase 6.
EOF

cat > quality/reports/latest.json <<EOF
{
  "status": "PHASE1",
  "coverage": ${GLOBAL:-null},
  "coverage_minimum": ${GLOBAL_MIN},
  "changed_code_minimum": ${CHANGED_MIN},
  "mutation_score": null,
  "security_issues": null,
  "notes": "Bootstrap report — expand in Phase 6"
}
EOF

echo "Report written to quality/reports/latest.md and quality/reports/latest.json"
