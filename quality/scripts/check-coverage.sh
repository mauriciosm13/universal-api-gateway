#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

ANALYZE_ONLY=0
for arg in "$@"; do
  case "$arg" in
    --analyze-only) ANALYZE_ONLY=1 ;;
  esac
done

COVERPROFILE="${ROOT}/coverage.out"
DIFF_FILE="${ROOT}/quality/reports/changed.diff"

# Packages with no tests fail `go test -coverprofile` on Go 1.25 (`no such tool "covdata"`).
# Include TestGoFiles and XTestGoFiles so external test packages still count.
test_packages() {
  go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./... | sed '/^$/d'
}

if [[ "$ANALYZE_ONLY" -eq 0 ]]; then
  echo "Running tests with coverage profile..."
  TEST_PKGS=()
  while IFS= read -r pkg; do
    TEST_PKGS+=("$pkg")
  done < <(test_packages)
  if [[ "${#TEST_PKGS[@]}" -eq 0 ]]; then
    echo "FAIL: no test packages found"
    exit 1
  fi
  go test -race -count=1 -covermode=atomic -coverprofile="$COVERPROFILE" "${TEST_PKGS[@]}"
elif [[ ! -f "$COVERPROFILE" ]]; then
  echo "FAIL: coverage profile missing at $COVERPROFILE"
  exit 1
fi

mkdir -p "${ROOT}/quality/reports"

DIFF_OK=0
REQUIRE_DIFF=0
if [[ "${GITHUB_EVENT_NAME:-}" == "pull_request" ]]; then
  REQUIRE_DIFF=1
fi

BASE_REF="${GITHUB_BASE_REF:-main}"
COMPARE_REF="origin/${BASE_REF}"

if git rev-parse --verify "$COMPARE_REF" >/dev/null 2>&1 \
  && git merge-base "$COMPARE_REF" HEAD >/dev/null 2>&1; then
  MERGE_BASE="$(git merge-base "$COMPARE_REF" HEAD)"
  # Compare merge-base to the working tree so uncommitted local edits are included.
  git diff -U0 "$MERGE_BASE" -- internal/ > "$DIFF_FILE"
  DIFF_OK=1
else
  : > "$DIFF_FILE"
  echo "Compare ref ${COMPARE_REF} not available for three-dot diff"
fi

ARGS=(
  -profile "$COVERPROFILE"
  -config "${ROOT}/quality/config/coverage.yaml"
  -baseline "${ROOT}/quality/baselines/quality-baseline.json"
  -diff "$DIFF_FILE"
  -report "${ROOT}/quality/reports/coverage.json"
)

if [[ "$DIFF_OK" -eq 1 ]]; then
  ARGS+=(-diff-ok)
fi
if [[ "$REQUIRE_DIFF" -eq 1 ]]; then
  ARGS+=(-require-diff)
fi

go run ./quality/covercheck/cmd/covercheck "${ARGS[@]}"
