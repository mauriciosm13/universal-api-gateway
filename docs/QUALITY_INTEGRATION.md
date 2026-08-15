# Quality Layer Integration

## Phase 1 (complete)

1. `quality/` at repository root — config, baselines, scripts, checklists
2. CI test job runs `quality/scripts/quality-gate.sh`
3. `make quality` for local runs
4. Engineering Intelligence principle in [PROJECT_BIBLE.md](PROJECT_BIBLE.md)
5. Quality-gate rules in [AGENTS.md](../AGENTS.md)
6. Quality checklist merged into [REVIEW_CHECKLIST.md](REVIEW_CHECKLIST.md)
7. Baseline in `quality/baselines/quality-baseline.json`
8. Coverage hardening (RFC 0010): single test run, regression, untested-package, changed-statement gates

## Later phases

Implement each remaining gate through RFC/ADR decisions. Coverage gates are hard after RFC 0010.

Every new CI dependency must have a documented purpose and, when architecturally significant, an ADR.
