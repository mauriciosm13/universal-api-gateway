# Quality Review Checklist

## Tests
- [ ] Unit tests
- [ ] Integration tests (when applicable)
- [ ] Regression test for every bug fix
- [ ] E2E tests for critical paths (when applicable)
- [ ] Mutation score acceptable (when gate active)

## Security
- [ ] No secrets committed
- [ ] Inputs validated
- [ ] Security scan reviewed (when gate active)

## Architecture
- [ ] Dependency direction preserved (hexagonal boundaries)
- [ ] No unnecessary dependency (ADR when significant)
- [ ] No global mutable state

## Performance
- [ ] Benchmark exists where relevant
- [ ] No material latency regression (when gate active)
- [ ] No material memory regression (when gate active)

## Maintainability
- [ ] Complexity acceptable
- [ ] Modules cohesive
- [ ] Documentation updated

## Quality gate
- [ ] `make quality` passes locally
- [ ] Coverage thresholds respected
