#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

mkdir -p quality/reports

COVER_JSON="quality/reports/coverage.json"

GLOBAL="n/a"
GLOBAL_MIN="85"
GLOBAL_STATUS="n/a"
CHANGED_STATUS="n/a"
REGRESSION_STATUS="n/a"
UNTESTED="none"
COVER_STATUS="n/a"

if [[ -f "$COVER_JSON" ]]; then
  GLOBAL="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("global","n/a"))' "$COVER_JSON")"
  GLOBAL_MIN="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("global_minimum","85"))' "$COVER_JSON")"
  GLOBAL_STATUS="$(python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); print("PASS" if d.get("global_ok") else "FAIL")' "$COVER_JSON")"
  CHANGED_STATUS="$(python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); print("SKIP" if d.get("changed_skipped") else ("PASS" if d.get("changed_ok") else "FAIL"))' "$COVER_JSON")"
  REGRESSION_STATUS="$(python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); print("PASS" if d.get("regression_ok") else "FAIL")' "$COVER_JSON")"
  UNTESTED="$(python3 -c 'import json,sys; pkgs=json.load(open(sys.argv[1])).get("untested_packages") or []; print("none" if not pkgs else ", ".join(pkgs))' "$COVER_JSON")"
  COVER_STATUS="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("status","n/a"))' "$COVER_JSON")"
elif [[ -f coverage.out ]]; then
  GLOBAL="$(go tool cover -func=coverage.out | awk '/^total:/ {gsub(/%/,"",$3); print $3}')"
fi

cat > quality/reports/latest.md <<EOF
# Quality Report

Status: PHASE1

| Gate | Status |
|---|---|
| Formatting | PASS (when gate completes) |
| Static analysis | PASS (when gate completes) |
| Unit tests | PASS (when gate completes) |
| Coverage overall | ${COVER_STATUS} |
| Global coverage | ${GLOBAL}% / ${GLOBAL_MIN}% minimum (${GLOBAL_STATUS}) |
| Coverage regression | ${REGRESSION_STATUS} |
| Changed-code coverage | ${CHANGED_STATUS} |
| Untested executable packages | ${UNTESTED} |
| Mutation | TODO Phase 3 |
| Integration / E2E | TODO Phase 2 |
| Security | TODO Phase 4 |
| Complexity / size | TODO Phase 3 |
| API / performance | TODO Phase 5 |

Full machine-readable coverage report: \`quality/reports/coverage.json\`.
EOF

GLOBAL_JSON="null"
if [[ "$GLOBAL" != "n/a" ]]; then
  GLOBAL_JSON="$GLOBAL"
fi

CHANGED_MIN="$(awk '/changed_code_minimum:/ {print $2}' quality/config/coverage.yaml)"

cat > quality/reports/latest.json <<EOF
{
  "status": "PHASE1",
  "coverage": ${GLOBAL_JSON},
  "coverage_minimum": ${GLOBAL_MIN},
  "changed_code_minimum": ${CHANGED_MIN},
  "coverage_gate": "${COVER_STATUS}",
  "mutation_score": null,
  "security_issues": null,
  "notes": "Coverage details in coverage.json — expand remaining gates in Phase 6"
}
EOF

echo "Report written to quality/reports/latest.md and quality/reports/latest.json"
