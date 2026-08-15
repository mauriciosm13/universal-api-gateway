# Universal API Gateway — Engineering Intelligence & Quality Layer

Objective engineering gates for AI-assisted development.

## Core principle

> AI can propose code. Automated engineering gates decide whether the code is acceptable.

## Scope

- Test coverage (Phase 1: global hard, regression hard, untested-package hard, changed-code hard)
- Mutation testing (Phase 3)
- Regression testing (Phase 2)
- Integration and E2E testing (Phase 2)
- Dependency health (Phase 4)
- Complexity and module size (Phase 3)
- Security scanning (Phase 4)
- API compatibility (Phase 5)
- Performance regression detection (Phase 5)
- Machine-readable quality reports (Phase 6)
- AI-assisted remediation protocol (Phase 6 automation)

Hard gates block merge. Soft gates warn while the project establishes baselines. Coverage gates are hard after RFC 0010.

## Usage

```bash
make quality          # full gate locally
bash quality/scripts/quality-gate.sh
```

See [QUALITY_GATE.md](QUALITY_GATE.md), [AI_REMEDIATION.md](AI_REMEDIATION.md), and [docs/engineering-intelligence.md](../docs/engineering-intelligence.md).
