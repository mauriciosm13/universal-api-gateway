# Engineering Intelligence Layer

AI agents are first-class development participants, but they are not the final authority on code quality.

```text
AI Agent
   |
   v
Code Change
   |
   v
Quality Gate
   +--> PASS --> Review/Merge
   |
   +--> FAIL --> Quality Report --> AI Remediation
                                      +--> retry (max 3)
                                      +--> Human Review
```

## Core principle

> AI can propose code. Automated engineering gates decide whether the code is acceptable.

## Scope

The layer evaluates formatting, static analysis, tests, coverage, mutation score, integration and E2E tests, dependency health, complexity, module size, security, API compatibility, and performance regression.

Implementation lives under `quality/` and CI — not inside the gateway runtime.

## Phases

| Phase | Focus |
|---|---|
| 1 | Format, vet, unit tests, global coverage, regression, changed-code, baseline, report bootstrap |
| 2 | Integration, regression, E2E |
| 3 | Mutation, complexity, module size |
| 4 | Dependency security, SAST, secrets, containers, SBOM |
| 5 | OpenAPI compatibility, performance benchmarks |
| 6 | Machine-readable reports, AI remediation workflow |

See [RFC 0006](rfcs/0006-engineering-intelligence-layer.md) and [QUALITY_INTEGRATION.md](QUALITY_INTEGRATION.md).
