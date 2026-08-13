# MVP Scope — M0 through M3

> **Goal:** Production-usable API Gateway with routing, authentication, and traffic control.  
> **Target timeline:** ~8 weeks (accelerated track)  
> **Rule:** Everything in this document is in scope for MVP. Everything else waits until post-MVP.

---

## MVP Definition

The MVP is **not** the full product vision (M4–M12). It is a gateway that can:

1. Route HTTP traffic by path, header, host, and method
2. Authenticate requests (JWT + API keys)
3. Control traffic (rate limiting, timeouts, retries, basic load balancing)
4. Run in Docker / Kubernetes with env-based configuration
5. Pass `make quality` and ship with RFC-lite + tests per deliverable

---

## Milestone Status at MVP Start

| Milestone | MVP status |
|---|---|
| M0 — Foundation | Complete |
| M1 — Core Gateway | ~65% — in progress |
| M2 — Auth & Authorization | Not started (MVP subset only) |
| M3 — Traffic Management | Not started (MVP subset only) |

---

## In Scope

### M1 — Core Gateway (finish)

| Deliverable | Status | Notes |
|---|---|---|
| Reverse proxy | Done | `GATEWAY_DEFAULT_UPSTREAM` |
| Path routing | Done | `GATEWAY_PATH_ROUTES` |
| Header routing | Done | `GATEWAY_HEADER_ROUTES` |
| Centralized route registration | Done | `internal/server/routes.go` |
| JSON error responses | Done | Gateway 4xx/5xx |
| Host routing | Done | `GATEWAY_HOST_ROUTES` |
| Method routing | Done | `GATEWAY_METHOD_ROUTES` |
| Request middleware pipeline | Done | Week 2 |
| Error middleware | Done | Week 2 |
| Environment variables | Done | MVP config model |

### M2 — Auth (MVP subset)

| Deliverable | Status | Notes |
|---|---|---|
| JWT validation | Planned | Week 3 |
| API keys | Planned | Week 4 |
| User context propagation | Planned | Week 4 |
| OAuth2 / OIDC | Post-MVP | |
| mTLS | Post-MVP | |
| RBAC / ABAC | Post-MVP | |

### M3 — Traffic (MVP subset)

| Deliverable | Status | Notes |
|---|---|---|
| Rate limiting (token bucket) | Planned | Week 5 |
| Request timeout | Planned | Week 6 |
| Retry (idempotent methods) | Planned | Week 6 |
| Load balancing (round robin) | Planned | Week 7 |
| Circuit breaker | Post-MVP | |
| Redis cache | Post-MVP | |
| Sticky sessions | Post-MVP | |

### Infrastructure (MVP)

| Deliverable | Status | Notes |
|---|---|---|
| Docker / Compose | Done | |
| K8s base manifests | Done | dev overlay sufficient for MVP |
| CI quality gate | Done | Phase 1 |
| Health endpoints | Done | |

---

## Out of Scope (Post-MVP)

Do **not** implement these during the MVP track unless the MVP scope doc is explicitly updated.

### M1 deferrals

- Regex routing
- URL rewrite / strip-prefix
- Request forwarding rules (beyond reverse proxy)
- Response transformation
- YAML configuration
- Hot reload

### M2 deferrals

- OAuth2, OpenID Connect
- Basic authentication
- Mutual TLS
- RBAC, ABAC, policies

### M3 deferrals

- Sliding window / leaky bucket rate limiting (token bucket first)
- Least connections / weighted / sticky load balancing
- Circuit breaker, bulkhead, fallback
- Memory / Redis caching

### M4–M12

- Full observability stack (Prometheus, Grafana, trace instrumentation)
- Plugin platform (WASM / Lua)
- API Management / Developer Portal
- Multi-cloud adapters
- Edge platform, AI gateway, enterprise features

---

## Accelerated Process Rules

To hit the 8-week target, use **RFC-lite** for MVP:

1. **One RFC per milestone phase** (optional) or one combined `docs/rfcs/MVP-m1-m3.md` — not one RFC per small feature
2. **Spec required** — use `docs/templates/SPEC.md`, keep it short
3. **ADR only** when adding a new dependency
4. **Tests required** — unit + at least one integration test per deliverable
5. **Batch PRs** — up to 2 related features per PR when they share wiring
6. **`make quality` must pass** — non-negotiable

---

## 8-Week Backlog

Assumes ~1 focused week per row. Adjust if part-time.

### Week 1 — Routing completion

- [x] Host routing (`GATEWAY_HOST_ROUTES`)
- [x] Method routing (`GATEWAY_METHOD_ROUTES`)
- [x] Router chain precedence updated: host → header → method → path → default
- [x] Tests + spec updates

**Exit criteria:** Request routed by host and HTTP method.

### Week 2 — Middleware pipeline

- [x] Request middleware chain (ordered execution before proxy)
- [x] Error middleware (consistent JSON errors, central handler)
- [x] Wire auth and rate-limit hooks as no-op passthrough until M2/M3
- [x] Tests + spec

**Exit criteria:** Middleware runs on all proxied traffic; errors flow through one path.

### Week 3 — JWT authentication

- [ ] JWT validator adapter (`Authorization: Bearer`)
- [ ] Config: JWKS URL or HMAC secret via env
- [ ] Reject invalid/expired tokens with JSON 401
- [ ] Tests + spec

**Exit criteria:** Valid JWT passes; invalid JWT returns 401 JSON.

### Week 4 — API keys + identity

- [ ] API key validation (header or query, config via env)
- [ ] User context attached to request (claims / key identity)
- [ ] Auth middleware integrated in pipeline
- [ ] Tests + spec

**Exit criteria:** Request authenticated by JWT or API key; identity available to downstream routing.

### Week 5 — Rate limiting

- [ ] Token bucket limiter adapter
- [ ] Config: limits per route or global default
- [ ] JSON 429 response
- [ ] Tests + spec

**Exit criteria:** Excess traffic receives 429; within limit passes through.

### Week 6 — Timeout and retry

- [ ] Upstream request timeout (configurable)
- [ ] Retry for idempotent methods (GET, HEAD, OPTIONS) with max attempts
- [ ] JSON 504/502 where appropriate
- [ ] Tests + spec

**Exit criteria:** Slow upstream times out; transient failures retried safely.

### Week 7 — Load balancing

- [ ] Multiple upstreams per route (comma-separated URLs)
- [ ] Round robin selection
- [ ] Health-aware skip deferred — simple round robin only
- [ ] Tests + spec

**Exit criteria:** Traffic distributed across configured upstreams.

### Week 8 — MVP polish and release

- [ ] End-to-end demo: Docker Compose with 2 upstreams + auth + rate limit
- [ ] Update README, DEPLOYMENT, CHANGELOG
- [ ] Mark M1–M3 MVP items done in ROADMAP
- [ ] Known limitations doc updated
- [ ] Tag `v0.2.0-mvp` (or semver per team preference)

**Exit criteria:** New engineer can run MVP stack locally in under 15 minutes using docs.

---

## Success Criteria (MVP Done)

- [ ] All **In Scope** items checked
- [ ] No **Out of Scope** work merged without scope doc update
- [ ] `make quality` passes on `main`
- [ ] Gateway deployable via Docker Compose and K8s dev overlay
- [ ] Demo script or documented curl flow for: route → auth → rate limit → upstream

---

## References

- [ROADMAP.md](ROADMAP.md) — full product vision
- [PROJECT_BIBLE.md](PROJECT_BIBLE.md) — engineering principles
- [DEPLOYMENT.md](DEPLOYMENT.md) — run targets
