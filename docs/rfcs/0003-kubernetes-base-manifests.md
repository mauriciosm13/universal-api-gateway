# RFC 0003 — Kubernetes Base Manifests (M0)

## Summary

Add cloud-agnostic Kubernetes manifests under `deploy/kubernetes/` using Kustomize: a shared base (Deployment, Service, ConfigMap) and environment overlays (`dev`, `staging`, `prod`). Fixed namespace `gateway`, health probes on existing endpoints, configuration via ConfigMap.

## Motivation

Milestone 0 requires a deployable foundation beyond Docker. The gateway already exposes liveness and readiness probes and loads configuration from environment variables. Kubernetes manifests close the M0 deployment gap without waiting for M8 production orchestration (Ingress, HPA, Helm, CD).

## Goals

- Base manifests: Deployment, Service, ConfigMap with liveness/readiness probes
- Kustomize overlays for dev, staging, and prod (replicas, resources, image tags)
- Fixed namespace `gateway` across all overlays
- Document apply/build workflow in DEPLOYMENT.md
- CI validation via `kubectl kustomize`

## Non Goals

- Ingress, TLS, Gateway API (M8)
- HPA, PDB, NetworkPolicy (M8)
- Helm chart (M8)
- Secrets management (ConfigMap only for M0)
- Image registry publish or CD pipeline (M8)
- Cloud-specific annotations (AWS, GCP, Azure)

## Detailed Design

### Directory layout

```text
deploy/kubernetes/
  base/
    kustomization.yaml
    deployment.yaml
    service.yaml
    configmap.yaml
  overlays/
    dev/
      kustomization.yaml
      patch-deployment.yaml
    staging/
      kustomization.yaml
      patch-deployment.yaml
    prod/
      kustomization.yaml
      patch-deployment.yaml
```

### Base resources

**Deployment**

- Single container on port 8080
- Image: `universal-api-gateway:latest` (patched per overlay)
- `envFrom.configMapRef`: `gateway-config`
- Liveness: `GET /health/live` on port 8080
- Readiness: `GET /health/ready` on port 8080
- Security: non-root, read-only root filesystem, no privilege escalation
- Modest default resource requests/limits

**Service**

- ClusterIP, port 8080 → targetPort 8080
- Selector matches Deployment labels

**ConfigMap**

- `GATEWAY_HOST`, `GATEWAY_PORT`, `GATEWAY_READ_TIMEOUT`, `GATEWAY_WRITE_TIMEOUT`, `GATEWAY_IDLE_TIMEOUT`
- Values mirror `internal/config` defaults

**Namespace**

- Fixed: `gateway` (set in base `kustomization.yaml`)

### Overlays

Overlays reference `../../base` and patch only environment-specific fields:

| Overlay | Replicas | Image tag | Resources |
|---|---|---|---|
| dev | 1 | `:latest` | minimum |
| staging | 2 | `:staging` | base |
| prod | 3 | semver tag | increased |

No per-overlay namespace — one cluster runs one overlay at a time in `gateway`.

### CI

Add a CI step that runs `kubectl kustomize` on base and each overlay. Invalid YAML or Kustomize errors fail the build.

## Alternatives

1. **Plain YAML per environment** — rejected; duplicates Deployment/Service/ConfigMap three times.
2. **Helm only** — rejected for M0; heavier tooling, deferred to M8.
3. **Kustomize base + overlays (chosen)** — built into kubectl, DRY, no new Go dependencies.

## Risks

- **Image not in cluster** — manifests assume image is built locally or pushed separately. Documented in DEPLOYMENT.md.
- **ConfigMap vs Secret** — sensitive values not in M0 scope; ConfigMap is sufficient for non-secret gateway settings.
- **Same namespace across overlays** — cannot run dev and staging concurrently in one cluster; acceptable for M0.

## Migration

No application code changes. New files under `deploy/kubernetes/`. Docker and Docker Compose unchanged.

## Open Questions

- None for M0. Ingress and HPA tracked under M8.
