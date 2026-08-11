# AI Quality Remediation Protocol

1. Read [PROJECT_BIBLE.md](../docs/PROJECT_BIBLE.md).
2. Read [ENGINEERING_MANUAL.md](../docs/ENGINEERING_MANUAL.md).
3. Read relevant RFC/ADR/spec.
4. Read the complete quality report.
5. Identify root cause.
6. Make the smallest safe change.
7. Run affected tests.
8. Run the full quality gate (`make quality`).
9. Report what changed and why.

## Forbidden actions

Never lower thresholds, remove tests, weaken assertions, suppress security findings, disable checks, remove observability, remove benchmarks, modify baselines to pass, or change public contracts silently.

Maximum automated repair attempts: 3. Then stop and request human review.
