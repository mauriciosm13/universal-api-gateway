#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

CONFIG="${ROOT}/quality/config/coverage.yaml"
COVERPROFILE="${ROOT}/coverage.out"
ANALYZE_ONLY=0

for arg in "$@"; do
  case "$arg" in
    --analyze-only) ANALYZE_ONLY=1 ;;
  esac
done

GLOBAL_MIN="$(awk '/global_minimum:/ {print $2}' "$CONFIG")"
CHANGED_MIN="$(awk '/changed_code_minimum:/ {print $2}' "$CONFIG")"

if [[ -z "$GLOBAL_MIN" || -z "$CHANGED_MIN" ]]; then
  echo "FAIL: could not read coverage thresholds from $CONFIG"
  exit 1
fi

collect_coverage() {
  local profile="$1"
  rm -f "$profile" "${ROOT}/tmp.out"

  while IFS= read -r pkg; do
    [[ -z "$pkg" ]] && continue
    go test -race -count=1 -coverprofile="${ROOT}/tmp.out" "$pkg"
    if [[ ! -s "$profile" ]]; then
      cp "${ROOT}/tmp.out" "$profile"
    else
      tail -n +2 "${ROOT}/tmp.out" >> "$profile"
    fi
  done < <(go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./internal/...)

  rm -f "${ROOT}/tmp.out"
}

if [[ "$ANALYZE_ONLY" -eq 0 ]]; then
  echo "Running tests with coverage profile..."
  collect_coverage "$COVERPROFILE"
elif [[ ! -f "$COVERPROFILE" ]]; then
  echo "FAIL: coverage profile missing at $COVERPROFILE"
  exit 1
fi

GLOBAL="$(go tool cover -func="$COVERPROFILE" | awk '/^total:/ {gsub(/%/,"",$3); print $3}')"
if [[ -z "$GLOBAL" ]]; then
  echo "FAIL: could not parse global coverage"
  exit 1
fi

echo "Global coverage: ${GLOBAL}% (minimum ${GLOBAL_MIN}%)"

GLOBAL_OK="$(awk -v g="$GLOBAL" -v m="$GLOBAL_MIN" 'BEGIN { print (g >= m) ? "yes" : "no" }')"
if [[ "$GLOBAL_OK" != "yes" ]]; then
  echo "FAIL: global coverage ${GLOBAL}% is below minimum ${GLOBAL_MIN}%"
  exit 1
fi

echo "PASS: global coverage"

BASE_REF="${GITHUB_BASE_REF:-main}"
COMPARE_REF="origin/${BASE_REF}"

if git rev-parse --verify "$COMPARE_REF" >/dev/null 2>&1; then
  CHANGED_DIRS=""
  CHANGED_DIRS="$(git diff --name-only "${COMPARE_REF}...HEAD" -- 'internal/**/*.go' \
    | grep -v '_test\.go$' \
    | xargs -I{} dirname {} 2>/dev/null \
    | sort -u \
    || true)"

  if [[ -n "$CHANGED_DIRS" ]]; then
    echo "Changed packages (soft gate, minimum ${CHANGED_MIN}%):"
    WARN=0
    while IFS= read -r dir; do
      [[ -z "$dir" ]] && continue
      PKG_PROFILE="$(mktemp)"
      PKG_COV="$(go test -count=1 -coverprofile="$PKG_PROFILE" "./${dir}/..." >/dev/null 2>&1 \
        && go tool cover -func="$PKG_PROFILE" | awk '/^total:/ {gsub(/%/,"",$3); print $3}')"
      rm -f "$PKG_PROFILE"
      if [[ -z "$PKG_COV" ]]; then
        PKG_COV="0.0"
      fi
      echo "  ${dir}: ${PKG_COV}%"
      PKG_OK="$(awk -v c="$PKG_COV" -v m="$CHANGED_MIN" 'BEGIN { print (c >= m) ? "yes" : "no" }')"
      if [[ "$PKG_OK" != "yes" ]]; then
        WARN=1
      fi
    done <<< "$CHANGED_DIRS"

    if [[ "$WARN" -eq 1 ]]; then
      echo "WARN: one or more changed packages are below ${CHANGED_MIN}% (soft gate in Phase 1)"
    else
      echo "PASS: changed packages meet ${CHANGED_MIN}%"
    fi
  else
    echo "No changed Go packages under internal/ — skipping changed-package check"
  fi
else
  echo "Compare ref ${COMPARE_REF} not found — skipping changed-package check"
fi
