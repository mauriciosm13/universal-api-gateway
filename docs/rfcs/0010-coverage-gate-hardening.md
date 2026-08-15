# RFC 0010 — Coverage Gate Hardening

## Summary

Harden the Phase 1 coverage gate so the number matches executable `internal/` code, CI runs tests once, baseline regression is enforced, and changed code is measured at statement granularity.

## Motivation

RFC 0006 shipped a working coverage script with known gaps: packages without tests were omitted from the denominator, `fail_on_regression` was unused, changed-package checks were soft and coarse, tests ran twice, and shallow clones skipped the PR diff.

## Goals

- One `go test -race -covermode=atomic -coverprofile` per `make quality` / CI test job, limited to packages that have tests (Go 1.25 `covdata` fails coverprofile on packages with no tests)
- Global coverage counts `internal/` statements from that profile
- Executable `internal/` packages without tests fail the gate
- Types-only packages (ports, structs) stay exempt
- `fail_on_regression` compares against `quality/baselines/quality-baseline.json`
- Changed-code coverage is added coverable statements vs the PR base (90% hard on pull requests)
- Full git history in CI so the three-dot diff can run
- Analyzer lives under `quality/covercheck` with unit tests (no new module dependency)

## Non Goals

- Third-party coverage SaaS or `diff-cover` binaries (no new CI dependency; revisit via ADR if needed)
- Covering `cmd/` mains
- HTML/Cobertura PR comments
- Weakening the 85% global minimum

## Detailed Design

`quality/scripts/quality-gate.sh` runs `quality/scripts/check-coverage.sh`, which tests every package that has `TestGoFiles` or `XTestGoFiles` once and writes `coverage.out`. It then invokes `go run ./quality/covercheck/cmd/covercheck`.

Changed-code coverage maps `git diff -U0 <merge-base>` (working tree, so uncommitted edits count) added lines onto cover-profile blocks. Lines with no block (comments, type declarations) are ignored. Pull requests fail if the base-ref diff cannot be produced.

Re-baselining `measured_global` is allowed only when the metric definition changes, and must be documented. Lowering `global_minimum` remains forbidden.

## Alternatives

1. **`-coverpkg=./internal/...` on every package** — rejected; `go test` fails with `go: no such tool "covdata"` on packages without tests.
2. **`go test ./... -coverprofile`** — rejected for the same `covdata` failure on `cmd/` and other test-less packages.
3. **Keep package-level 90% WARN** — rejected; too coarse and did not block merge.
4. **External diff-cover tool** — deferred; stdlib-only analyzer is enough for statement overlap.

## Risks

- First run after the metric change must update `measured_global` to the new honest number, not the old 96.2% figure from a different denominator.
- Local `make quality` skips changed-code when `origin/main` is missing; CI pull requests require the diff.
