# Engineering Manual

Index of engineering documentation for **universal-api-gateway**.

## Product

- [Project Bible](PROJECT_BIBLE.md)
- [Roadmap](ROADMAP.md)
- [Architecture](ARCHITECTURE.md)

## Engineering

- [Coding Standard](CODING_STANDARD.md)
- [Anti-patterns](ANTI_PATTERNS.md)
- [Git Workflow](GIT_WORKFLOW.md)
- [Testing](TESTING.md)
- [Security](SECURITY.md)
- [Performance](PERFORMANCE.md)
- [Observability](OBSERVABILITY.md)
- [Deployment](DEPLOYMENT.md)
- [Review Checklist](REVIEW_CHECKLIST.md)

## Process Templates

- [RFC Template](templates/RFC.md)
- [ADR Template](templates/ADR.md)
- [Spec Template](templates/SPEC.md)
- [Task Template](templates/TASK.md)

## AI Workflow

See [AGENTS.md](../AGENTS.md) and `.ai/` for prompts, playbooks, and checklists.

## Local Development

```bash
make help    # available targets
make test    # run tests
make run     # start gateway on :8080
```

Dev Container: `.devcontainer/devcontainer.json` (Go 1.25, port 8080).
