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
| unit tests (`-race`, `-covermode=atomic`) | hard fail |
| global coverage (`internal/` statements) | hard fail at 85% |
| coverage regression vs baseline | hard fail (`allowed_drop_percentage: 0`) |
| executable `internal/` package without tests | hard fail |
| changed-code coverage on PRs | hard fail at 90% of added coverable statements |

`make quality` runs unit tests once (packages with tests, including `package x_test` files), writes `coverage.out`, then `quality/covercheck` analyzes that profile. No second test run. Go 1.25 cannot emit a coverprofile for packages with no tests (`covdata`).

Configuration: `quality/config/coverage.yaml`. Baseline: `quality/baselines/quality-baseline.json`.

Reports: `quality/reports/latest.md`, `quality/reports/latest.json`, and `quality/reports/coverage.json`.

See [engineering-intelligence.md](engineering-intelligence.md) and [RFC 0006](rfcs/0006-engineering-intelligence-layer.md).

## Planned gates

Mutation testing, integration/E2E infrastructure, security scanning, complexity/module size, OpenAPI compatibility, and performance regression detection roll out in later phases through RFC/ADR decisions.
