# RFC 0006 — Engineering Intelligence Layer (Phase 1)

## Summary

Introduce a CI-local **Engineering Intelligence and Quality Gate** layer under `quality/`. Phase 1 delivers formatting, static analysis, unit tests, global coverage enforcement, baseline tracking, and bootstrap placeholders for later gates. AI agents may propose code; automated gates decide merge eligibility.

## Motivation

The project follows Staff Engineer practices and AI-assisted development. Without objective gates, AI-generated code is indistinguishable from human code in risk — but not in accountability. `TESTING.md` already plans a coverage gate; this RFC formalizes the full quality pipeline and implements the first slice.

## Goals (Phase 1)

- `quality/` directory with configuration, baselines, scripts, and checklists
- Single entry point: `quality/scripts/quality-gate.sh`
- CI and Makefile invoke the same gate (no duplicated logic)
- Hard gates: format, `go vet`, unit tests (`-race`), global coverage ≥ 85%
- Soft gate: changed-package coverage ≥ 90% on pull requests (warn until stable)
- Human and machine-readable bootstrap quality report
- Update PROJECT_BIBLE, AGENTS, TESTING, ROADMAP, CHANGELOG

## Non Goals (Phase 1)

- Mutation testing (Phase 3 — requires ADR for tool selection)
- Integration, E2E, regression test infrastructure (Phase 2)
- SAST, secret scanning, container scanning, SBOM (Phase 4)
- OpenAPI compatibility and benchmark regression (Phase 5)
- Full JSON quality report schema and AI remediation automation (Phase 6)
- Runtime quality code inside `internal/` (forbidden by design)

## Detailed Design

### Architecture

```text
quality/
  config/           Thresholds (YAML)
  baselines/        Known-good measurements
  scripts/          Gate orchestration and checks
  checklists/       PR review extensions
  reports/          Generated output (gitignored)
docs/
  engineering-intelligence.md
  QUALITY_INTEGRATION.md
```

The gateway runtime remains unchanged. Quality is CI, tooling, and documentation only.

### Pipeline (Phase 1 active steps)

```text
Format → Static Analysis → Unit Tests → Coverage → Report
```

Remaining steps emit bootstrap placeholders with explicit TODOs referencing this RFC and future ADRs.

### Coverage policy

| Metric | Threshold | Phase 1 enforcement |
|---|---|---|
| Global (`internal/...`) | 85% | FAIL |
| Changed packages (PR) | 90% | WARN |

Configuration: `quality/config/coverage.yaml`. Baseline: `quality/baselines/quality-baseline.json`.

Changed-package coverage avoids new CI dependencies. Line-level changed coverage may adopt a dedicated tool through a future ADR.

### CI integration

`.github/workflows/ci.yml` test job calls `bash quality/scripts/quality-gate.sh` instead of duplicating fmt/vet/test steps. Build and Kubernetes validation jobs remain separate.

### AI remediation (documented, not automated in Phase 1)

See `quality/AI_REMEDIATION.md`. Agents must never lower thresholds, delete tests, suppress security findings, or change public contracts silently. Maximum three automated repair attempts when Phase 6 lands.

## Alternatives

1. **Separate `quality.yml` workflow** — rejected; duplicates existing CI and diverges on Go version and flags.
2. **Third-party quality platform only** — rejected for M0; adds vendor lock-in before baselines exist.
3. **Implement all 15 dimensions in one PR** — rejected; violates incremental rollout and ADR policy.
4. **Single orchestrator script + extended CI (chosen)** — one source of truth, phased gates.

## Risks

- **Coverage gate fails existing code** — mitigated by adding targeted unit tests before enabling the 85% hard gate.
- **Changed-code heuristic too coarse** — package-level WARN in Phase 1; refine with ADR-backed tooling later.
- **Placeholder gates look “green”** — bootstrap scripts print explicit TODO; report status is `phase1`, not `pass-all`.

## Migration

1. Land `quality/` and documentation.
2. Point CI and Makefile at `quality-gate.sh`.
3. Establish baseline on `main` after first green run.
4. Promote soft gates to hard gates through RFC/ADR updates.

## Open Questions

- Mutation testing tool: evaluate `go-mutesting`, `gremlins`, and alternatives in Phase 3 ADR.
- Line-level changed coverage tool: evaluate in a follow-up ADR if package-level checks prove insufficient.
