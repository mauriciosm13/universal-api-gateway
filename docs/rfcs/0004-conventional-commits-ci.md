# RFC 0004 — Conventional Commits in CI

## Summary

Enforce [Conventional Commits](https://www.conventionalcommits.org/) in GitHub Actions: validate PR titles and commit messages on pull requests and pushes to `main`. Aligns CI with existing `docs/GIT_WORKFLOW.md` policy.

## Motivation

The project already documents Conventional Commits and Semantic Versioning, but enforcement is manual. Mixed commit history (`feat(di): ...` vs bare `fix`) hurts readability and blocks automated changelogs/releases later.

## Goals

- Fail CI when PR title or commits violate Conventional Commits
- Zero changes to application code
- No new runtime dependencies in the Go module
- Document format and examples in `docs/GIT_WORKFLOW.md`

## Non Goals

- Commit hooks (local Husky/commitlint) — optional future DX improvement
- Automated releases or changelog generation — enabled later once history is clean
- Rewriting existing git history

## Detailed Design

### Tooling

| Check | Tool | Trigger |
|---|---|---|
| Commits in PR / push to main | `wagoid/commitlint-github-action` + `@commitlint/config-conventional` | `pull_request`, `push` → `main` |
| PR title | `amannn/action-semantic-pull-request` | `pull_request` only |

Configuration file: `.commitlintrc.yaml` at repo root. Workflow: `.github/workflows/conventional-commits.yml`.

### Allowed types

`feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

Scope is **recommended** but not required (`feat(di): ...` or `feat: ...`).

### Example valid messages

```text
feat(k8s): add base manifests with kustomize overlays
fix: handle shutdown timeout on SIGTERM
docs: expand GIT_WORKFLOW with commit examples
ci: validate conventional commits in pull requests
```

### Example invalid messages

```text
fix
WIP
updated stuff
Feature: add routing
```

## Alternatives

1. **Document only** — rejected; already documented, not followed consistently.
2. **PR title only** — rejected; commits in PR can still be messy.
3. **Local hooks only** — rejected; not enforced for all contributors, easy to bypass with `--no-verify`.
4. **PR title + commitlint in CI (chosen)** — enforced at merge gate, no Go module deps.

## Risks

- **Legacy commits on open PRs** — may fail until rebased/squashed with proper messages.
- **False positives on merge commits** — mitigated by commitlint action defaults.

## Migration

New workflow only. Existing commits on `main` are not rewritten. New PRs must comply.

## Open Questions

- None for M0.
