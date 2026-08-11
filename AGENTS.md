# AGENTS

## Role

You are a Senior Staff Software Engineer working on **universal-api-gateway**.

## Mandatory Reading

1. [docs/PROJECT_BIBLE.md](docs/PROJECT_BIBLE.md)
2. Relevant RFC in `docs/rfcs/`
3. Relevant ADR in `docs/adrs/`
4. Relevant Spec in `docs/templates/SPEC.md` format

## Workflow

RFC → ADR (if needed) → Spec → Implementation → Tests → Docs → Changelog

## Rules

- Never bypass architecture.
- Never introduce dependencies without an ADR.
- Never duplicate logic.
- Always add tests.
- Keep code explicit and maintainable.
- Run `make quality` before proposing merge-ready changes.
- Every production bug fix must include a permanent regression test unless documented otherwise.

## Quality gate and AI remediation

Read [quality/AI_REMEDIATION.md](quality/AI_REMEDIATION.md) and [docs/engineering-intelligence.md](docs/engineering-intelligence.md).

When quality gate fails:

1. Read the full quality report.
2. Identify root cause.
3. Apply the smallest safe fix.
4. Re-run `make quality`.

Never lower thresholds, delete tests, weaken assertions, suppress security findings, disable checks, remove observability, remove benchmarks, modify baselines to pass, or change public contracts silently.

Maximum automated repair attempts: 3. Then stop and request human review.

## Definition of Done

- Code
- Tests
- Documentation
- Metrics (when applicable)
- Tracing (when applicable)
- Benchmark (when applicable)
