# MVP Execution Plan — Weeks 6–8 + AWS Lambda Deploy

> **Status:** Approved design — implementation ready  
> **Target release:** `v0.2.0-mvp`  
> **Duration:** 3 calendar weeks (parallel streams)

## Objective

Complete MVP scope (M1–M3 subset), deploy to AWS Lambda staging via AWS CLI, and ship a reproducible E2E demo.

## Design documents

| Document | Purpose |
|---|---|
| [mvp-weeks-6-8-reliability-lb-release.md](rfcs/mvp-weeks-6-8-reliability-lb-release.md) | Features Week 6–8 |
| [mvp-aws-lambda-deployment.md](rfcs/mvp-aws-lambda-deployment.md) | AWS Lambda deploy |
| [upstream-timeout-retry.md](specs/upstream-timeout-retry.md) | Week 6 spec |
| [load-balancing-round-robin.md](specs/load-balancing-round-robin.md) | Week 7 spec |
| [aws-lambda-runtime.md](specs/aws-lambda-runtime.md) | Lambda spec |
| [ADR 0003](adrs/0003-aws-lambda-runtime-adapter.md) | LWA decision |
| [ADR 0004](adrs/0004-rate-limit-lambda-strategy.md) | Rate limit on Lambda |
| [ADR 0005](adrs/0005-aws-deploy-aws-cli.md) | AWS CLI scripts |

---

## Parallel streams

```text
                    ┌─────────────────────────────────────┐
                    │  Week 1: Design ✅ + kickoff        │
                    └─────────────────────────────────────┘
           ┌────────┴────────┬────────────────┬───────────┴────────┐
           ▼                 ▼                ▼                    ▼
    Stream A            Stream B         Stream C              Stream D
    Timeout+Retry       Round Robin      AWS Lambda            Docs/E2E
    (Week 6)            (Week 7)         (Weeks 6–8)           (Week 8)
           │                 │                │                    │
           └────────┬────────┘                │                    │
                    ▼                         │                    │
              merge to main ◄─────────────────┘                    │
                    │                                                │
                    └──────────────────────► demo + tag v0.2.0-mvp ◄┘
```

| Stream | Owner focus | Depends on |
|---|---|---|
| **A** — Reliability | `internal/reliability/`, config, gateway transport | Spec approved |
| **B** — Load balancing | `Route.Upstreams`, round robin selector | Route struct defined (Day 1) |
| **C** — AWS deploy | `Dockerfile.lambda`, `deploy/aws/`, CI | ADRs approved |
| **D** — Release | compose demo, scripts, docs | A + B merged |

Streams **A** and **C** start **Day 1**. Stream **B** starts when `Route.Upstreams` interface is merged or stubbed (end of Week 1 / start Week 2).

---

## Week 1 — Foundation + parallel kickoff

### Days 1–2 (design — done)

- [x] RFC MVP-5 (weeks 6–8)
- [x] RFC MVP-6 (AWS Lambda)
- [x] Specs + ADRs 0003–0005
- [x] Execution plan (this doc)
- [x] Deploy skeleton (`deploy/aws/`, `Dockerfile.lambda`)

### Days 3–5

| ID | Task | Stream | PR label |
|---|---|---|---|
| W1-A1 | `internal/config/reliability.go` + tests | A | `feat/reliability-config` |
| W1-A2 | `internal/reliability/adapter/timeout` | A | same PR or follow-up |
| W1-A3 | `internal/reliability/adapter/retry` | A | `feat/reliability-retry` |
| W1-A4 | Wire transport in `gateway.go` + integration tests | A | `feat/reliability-wire` |
| W1-B1 | Extend `port.Route` with `Upstreams []string` | B | `feat/lb-route-model` |
| W1-C1 | `GATEWAY_RUNTIME` in config + main.go branch | C | `feat/lambda-runtime` |
| W1-C2 | `Dockerfile.lambda` build verified locally | C | `feat/lambda-dockerfile` |
| W1-C3 | `setup-ecr.sh`, `setup-iam.sh` | C | `feat/aws-setup` |
| W1-C4 | First staging deploy (health only) | C | `feat/aws-deploy-staging` |

**Week 1 exit criteria**

- [ ] Timeout + retry merged; 504/502 tests pass
- [ ] Lambda container runs locally: `curl localhost:8080/health/ready`
- [ ] Staging Function URL returns `/health`

---

## Week 2 — Load balancing + AWS hardening

| ID | Task | Stream | PR label |
|---|---|---|---|
| W2-B1 | Parse comma-separated upstreams in config | B | `feat/lb-config` |
| W2-B2 | `roundrobin` selector adapter | B | `feat/lb-roundrobin` |
| W2-B3 | Gateway integration + distribution test | B | `feat/lb-wire` |
| W2-C1 | `setup-secrets.sh` + env JSON templates | C | `feat/aws-secrets` |
| W2-C2 | `deploy-lambda.sh` full env + secrets merge | C | `feat/aws-deploy-full` |
| W2-C3 | `build-and-push.sh` + ECR push | C | same |
| W2-C4 | `.github/workflows/deploy-aws.yml` (OIDC) | C | `feat/aws-ci` |
| W2-C5 | Redeploy staging with auth + upstream | C | — |
| W2-C6 | `smoke-test.sh` | C | `feat/aws-smoke` |
| W2-D1 | Bootstrap integration test (compose up + curl) | D | `test/e2e-bootstrap` |

**Week 2 exit criteria**

- [ ] Round robin: 2 upstreams alternate in tests
- [ ] CI pushes image to ECR on tag
- [ ] Staging smoke: auth + proxy to upstream pass

---

## Week 3 — Release polish + production readiness

| ID | Task | Stream | PR label |
|---|---|---|---|
| W3-D1 | `docker-compose.demo.yml` + mock upstreams | D | `feat/demo-compose` |
| W3-D2 | `scripts/demo.sh` | D | `feat/demo-script` |
| W3-D3 | Update README, DEPLOYMENT, CHANGELOG | D | `docs/mvp-release` |
| W3-D4 | Update `known-limitations.md`, ROADMAP, MVP_SCOPE | D | same |
| W3-C1 | Production deploy runbook + `lambda-env.prod.json` | C | `docs/aws-prod` |
| W3-C2 | Manual prod deploy (workflow_dispatch) | C | — |
| W3-A1 | Lambda timeout validation hardening (if gaps) | A | `fix/lambda-validation` |
| W3-ALL | Tag `v0.2.0-mvp` | ALL | release |

**Week 3 exit criteria (MVP done)**

- [x] All MVP_SCOPE in-scope items checked
- [x] `make quality` passes on `main`
- [x] Demo runnable in < 15 min via README
- [x] Staging Lambda: route → auth → rate limit → upstream
- [ ] Tag `v0.2.0-mvp` published (post-merge)

---

## PR strategy

| Rule | Detail |
|---|---|
| Max 2 related features per PR | e.g. config + timeout adapter |
| Each PR | tests + `make quality` |
| Stream C PRs | can merge independently of A/B after runtime mode lands |
| Order | A1 → A4 before B3; C2 before C5 |

Suggested merge order:

1. `feat/lambda-runtime` + `feat/lambda-dockerfile`
2. `feat/reliability-*` (Week 6)
3. `feat/lb-*` (Week 7)
4. `feat/aws-*` (incremental)
5. `feat/demo-*` + docs
6. Release tag

---

## AWS prerequisites (operator checklist)

Before Week 1 deploy:

- [ ] AWS account + IAM user/role with admin or scoped permissions
- [ ] AWS CLI v2 installed and `aws configure` (or SSO)
- [ ] Docker installed locally
- [ ] GitHub repo: OIDC provider + IAM role for Actions (Week 2)
- [ ] Upstream URL reachable from AWS (public HTTPS for MVP)

Environment variables for scripts:

```bash
export AWS_REGION=us-east-1
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export UAG_PROJECT=universal-api-gateway
```

---

## Demo scenario (Week 8)

`scripts/demo.sh` exercises:

1. `GET /health` → 200
2. `GET /api/echo` without auth → 401
3. `GET /api/echo` with valid JWT → 200 + upstream body
4. Burst requests → 429
5. Multiple requests → alternate upstream (`X-Upstream-Id` header from mocks)

---

## Risk register

| Risk | Likelihood | Impact | Owner action |
|---|---|---|---|
| Lambda cold start too slow | Medium | Medium | 512MB+; document; provisioned concurrency later |
| Retry exceeds Lambda timeout | Medium | High | Startup validation (spec) |
| OIDC setup delays CI | Medium | Low | Manual deploy fallback Week 2 |
| Per-instance rate limit confusion | High | Low | ADR 0004 + docs |

---

## Post-MVP backlog (explicitly deferred)

- DynamoDB rate limit adapter
- Terraform port of `deploy/aws/scripts`
- API Gateway custom domain in front of Function URL
- VPC Lambda for private upstreams
- Integration tests Phase 2 full suite
- Prometheus / trace instrumentation (M4)

---

## References

- [MVP_SCOPE.md](MVP_SCOPE.md)
- [ROADMAP.md](ROADMAP.md)
- [DEPLOYMENT.md](DEPLOYMENT.md)
- [deploy/aws/README.md](../deploy/aws/README.md)
