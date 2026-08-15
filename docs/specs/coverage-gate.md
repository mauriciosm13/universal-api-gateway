# Specification — Coverage Gate

## Overview

Enforce honest statement coverage for `internal/` during `make quality` and CI. Analysis is tooling under `quality/covercheck`, not gateway runtime.

## Responsibilities

| Component | Responsibility |
|---|---|
| `quality/scripts/quality-gate.sh` | Format, vet, then `check-coverage.sh` |
| `quality/scripts/check-coverage.sh` | Test packages with tests, write profile, resolve base-ref diff, run covercheck |
| `quality/covercheck` | Parse profile, config, baseline, and unified diff; emit PASS/FAIL |
| `quality/config/coverage.yaml` | Thresholds |
| `quality/baselines/quality-baseline.json` | Measured global coverage for regression |

## API

```text
go test -race -count=1 -covermode=atomic -coverprofile=coverage.out <packages with tests>
go run ./quality/covercheck/cmd/covercheck \
  -profile coverage.out \
  -config quality/config/coverage.yaml \
  -baseline quality/baselines/quality-baseline.json \
  -diff quality/reports/changed.diff \
  -report quality/reports/coverage.json
```

## Data Flow

1. Format and `go vet` (hard).
2. Unit tests for packages with `TestGoFiles` or `XTestGoFiles` write `coverage.out`.
3. Covercheck keeps only `module/internal/` blocks.
4. Global percent is covered statements / total statements in that profile.
5. Executable packages with no tests and no profile blocks fail.
6. Global percent must be ≥ 85% and ≥ baseline minus allowed drop.
7. On pull requests, added coverable statements vs the merge-base working tree must be ≥ 90%.

## Errors

- Missing profile, config, or baseline: fail
- Global below minimum: fail
- Regression below baseline: fail
- Untested executable package: fail
- PR without a usable base-ref diff: fail
- Changed-code coverage below 90%: fail

## Metrics

Coverage itself is the metric. No runtime gateway metrics.

## Logs

Human lines on stdout (`PASS:` / `FAIL:` / `SKIP:`). JSON at `quality/reports/coverage.json`.

## Tracing

Not applicable.

## Security

No network. Reads git diff, coverage profile, and config only.

## Performance

One race-enabled test run per gate. Covercheck is a short `go run`.

## Testing

`quality/covercheck` unit tests cover profile filtering, diff parsing, global/changed/regression/untested gates, and types-only exemption.
