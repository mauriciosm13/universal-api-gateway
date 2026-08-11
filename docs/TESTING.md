# Testing

## Strategy

- Unit tests for core logic
- Integration tests for HTTP endpoints and adapters (Phase 2)
- Benchmarks for hot paths (routing, rate limit)
- Regression test for every production bug fix
- E2E black-box tests for critical gateway flows (Phase 2)

## Quality gate (Phase 1)

CI and local development run the same orchestrator:

```bash
make quality
```

Active gates:

| Check | Enforcement |
|---|---|
| `gofmt` | hard fail |
| `go vet` | hard fail |
| unit tests (`-race`) | hard fail |
| global coverage (`internal/...`) | hard fail at 85% |
| changed-package coverage on PRs | soft warn at 90% |

Configuration: `quality/config/coverage.yaml`. Baseline: `quality/baselines/quality-baseline.json`.

Reports: `quality/reports/latest.md` and `quality/reports/latest.json`.

See [engineering-intelligence.md](engineering-intelligence.md) and [RFC 0006](rfcs/0006-engineering-intelligence-layer.md).

## Planned gates

Mutation testing, integration/E2E infrastructure, security scanning, complexity/module size, OpenAPI compatibility, and performance regression detection roll out in later phases through RFC/ADR decisions.
