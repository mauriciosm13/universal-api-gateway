# Quality Gate

## Pipeline

Format → Static Analysis → Unit Tests → Coverage → Mutation → Integration → Regression → E2E → Dependencies → Security → Complexity → Module Size → API Compatibility → Performance → Report

Phase 1 implements the first four steps plus a bootstrap report. Remaining steps print explicit TODO placeholders.

Every gate must eventually emit:

- status: PASS | WARN | FAIL
- measured value
- threshold
- baseline when applicable
- actionable message

## Hard gates (Phase 1)

- build failure (via tests)
- failed tests
- formatting drift
- `go vet` findings
- global coverage below 85%
- coverage below baseline (`fail_on_regression`, `allowed_drop_percentage: 0`)
- executable `internal/` package with no tests
- changed-code coverage below 90% on pull requests (added coverable statements vs the base branch)

## Soft gates (Phase 1)

- none for coverage; remaining pipeline steps are TODO placeholders

## AI remediation

AI may repair failures, but must never weaken thresholds, delete tests, suppress security findings, or change public contracts silently.

Maximum automated remediation attempts: 3. Then require human review.

See [AI_REMEDIATION.md](AI_REMEDIATION.md).
