# Deployment

## Supported Targets

- Docker
- Docker Compose
- Kubernetes (Kustomize base + overlays)
- **AWS Lambda** (container + Lambda Web Adapter, AWS CLI scripts)
- GCP, Azure (planned)

## Docker

```bash
docker build -t universal-api-gateway .
docker run --rm -p 8080:8080 universal-api-gateway
```

## Docker Compose

```bash
docker compose up --build
```

## Kubernetes

Manifests live under `deploy/kubernetes/`. Namespace is fixed to `gateway`.

### Layout

```text
deploy/kubernetes/
  base/                 Deployment, Service, ConfigMap, Namespace
  overlays/
    dev/                1 replica, :latest
    staging/            2 replicas, :staging
    prod/               3 replicas, :0.1.0
```

### Prerequisites

- `kubectl` configured for your cluster
- Container image available to the cluster (build locally or push to a registry)

```bash
docker build -t universal-api-gateway:latest .
# For minikube/kind: load the image into the cluster
```

### Render manifests

```bash
kubectl kustomize deploy/kubernetes/overlays/dev
```

### Apply

```bash
kubectl apply -k deploy/kubernetes/overlays/dev
```

### Verify

```bash
kubectl get pods -n gateway
kubectl port-forward -n gateway svc/universal-api-gateway 8080:8080
curl http://localhost:8080/health/ready
```

### Overlays

| Overlay | Replicas | Image tag | Use case |
|---|---|---|---|
| `dev` | 1 | `:latest` | Local / development cluster |
| `staging` | 2 | `:staging` | Pre-production |
| `prod` | 3 | `:0.1.0` | Production (pin semver in overlay) |

Only one overlay should run per cluster — all use namespace `gateway`.

See [RFC 0003](rfcs/0003-kubernetes-base-manifests.md) and [Spec](specs/kubernetes-base-manifests.md) for details.

## AWS Lambda

Primary MVP deploy target. Managed via **AWS CLI scripts** (see [ADR 0005](adrs/0005-aws-deploy-aws-cli.md)).

### Architecture

```text
Client → Lambda Function URL → Lambda Web Adapter → gateway :8080 → upstream
```

### Prerequisites

- AWS CLI v2 configured (`aws sts get-caller-identity`)
- Docker (build `Dockerfile.lambda`)
- `jq`, `curl`

### Quick start (staging)

```bash
export AWS_REGION=us-east-1

./deploy/aws/scripts/setup-ecr.sh
./deploy/aws/scripts/setup-iam.sh --env staging
./deploy/aws/scripts/setup-secrets.sh --env staging

cp deploy/aws/config/lambda-env.staging.json deploy/aws/config/lambda-env.staging.local.json
# edit upstream and auth settings

./deploy/aws/scripts/build-and-push.sh --tag latest
./deploy/aws/scripts/deploy-lambda.sh --env staging --tag latest
./deploy/aws/scripts/smoke-test.sh --env staging
```

Full runbook: [deploy/aws/README.md](../deploy/aws/README.md).

### Lambda-specific configuration

| Variable | Notes |
|---|---|
| `GATEWAY_RUNTIME` | Set to `lambda` |
| `GATEWAY_UPSTREAM_TIMEOUT` | Must be ≤ Lambda timeout − 2s (default Lambda 30s → use 25s) |
| Secrets | `GATEWAY_JWT_HMAC_SECRET`, `GATEWAY_API_KEYS` from Secrets Manager at deploy |

### MVP limitations on Lambda

- Rate limit and round robin are **per Lambda instance** ([ADR 0004](adrs/0004-rate-limit-lambda-strategy.md))
- Upstreams must be **publicly reachable** (no VPC in MVP)
- Payload limit ~6 MB (sync invoke)

See [RFC MVP-6](rfcs/mvp-aws-lambda-deployment.md) and [Spec — AWS Lambda runtime](specs/aws-lambda-runtime.md).

## Demo stack (local E2E)

```bash
docker compose -f docker-compose.demo.yml up --build
./scripts/demo.sh
```

## Environment

All configuration is injected via environment variables. In Kubernetes, non-secret settings are provided by ConfigMap `gateway-config`. See [README](../README.md#configuration).

## Cloud Agnostic Rule

No cloud-specific SDK in core modules. Cloud integrations are adapters behind ports. Kubernetes manifests use standard resources without cloud-provider annotations.
