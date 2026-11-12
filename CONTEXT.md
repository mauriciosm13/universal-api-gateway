# Context

## Goal

Close one open gap from `REPORT.md` section 7: the layer import table in `CONVENTIONS.md` ("Layers") was partly inferred, because `docs/ARCHITECTURE.md` is stale (v0.1, no adapter row). Establish from the code what each layer actually imports.

## Requirements and constraints

- Read-only investigation. No code changes, no new features, no invented requirements.
- Every finding cites an existing file path.
- If the code settles the table, update only the "Layers" section of `CONVENTIONS.md`.
- If it stays ambiguous, record that in `REPORT.md` section 7 instead of adding a rule.

## Decisions

- Only `recon` is dispatched. No planner, spec, implementer or reviewer.
- `docs/ARCHITECTURE.md` is not edited in this task.

## Relevant files

- `CONVENTIONS.md` (Layers table), `REPORT.md` (sections 7 and 8)
- `docs/ARCHITECTURE.md`, `docs/ANTI_PATTERNS.md`
- `internal/**` (all packages), `cmd/gateway/main.go`

## Tooling

Go 1.25.0 is not on the default `PATH`. Use:
`PATH=/tmp/uag-go/go/bin:$PATH GOPATH=/tmp/uag-go/gopath GOCACHE=/tmp/uag-go/cache GOTOOLCHAIN=local`

## Acceptance criteria

- Each row of the Layers table is confirmed or corrected with the internal imports actually observed, each with a file path.
- Any import that breaks the documented direction is listed with its path.
- The outcome is reflected in `CONVENTIONS.md` (settled) or `REPORT.md` section 7 (still ambiguous).
