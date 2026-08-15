# ADR 0005 — AWS Deploy Tooling (AWS CLI)

## Context

MVP requires deploy to AWS Lambda (ECR + Lambda + IAM + Secrets Manager + Function URL). IaC options: Terraform, AWS SAM, AWS CDK, CloudFormation, or imperative **AWS CLI scripts**.

Team preference: start with AWS CLI (equivalent role to `gcloud` for AWS), avoid SAM/Terraform learning curve for MVP.

## Decision

Manage AWS MVP infrastructure via **versioned shell scripts** under `deploy/aws/scripts/` that call **AWS CLI** (`aws`).

- Scripts must be **idempotent** (create-or-update pattern)
- Configuration templates in `deploy/aws/config/`
- Runbook in `deploy/aws/README.md`
- CI uses same scripts via GitHub Actions + OIDC IAM role
- No SAM, Terraform, or CDK in MVP

Script inventory:

| Script | Purpose |
|---|---|
| `setup-ecr.sh` | ECR repository |
| `setup-iam.sh` | Lambda execution role + policies |
| `setup-secrets.sh` | Secrets Manager secrets per env |
| `build-and-push.sh` | Docker build `Dockerfile.lambda` + ECR push |
| `deploy-lambda.sh` | Create/update function, env, Function URL |
| `smoke-test.sh` | curl health + optional auth |
| `common.sh` | Shared vars, logging, idempotent helpers |

Environment selection via `--env staging|prod` and `deploy/aws/config/lambda-env.<env>.json`.

## Consequences

**Positive**

- Low tooling overhead; only `aws` CLI required locally
- Scripts readable without HCL/YAML framework knowledge
- Fast path to first staging deploy
- Aligns with “scripts in repo” pattern used by Makefile

**Negative**

- No declarative state file; drift if console edited manually
- Harder to review infra diffs than Terraform plan
- Migration to Terraform later requires one-time port

## Alternatives

1. **AWS SAM** — rejected; team chose AWS CLI; SAM is serverless-specific, not general AWS CLI.
2. **Terraform** — rejected for MVP; best post-MVP when multi-env drift matters.
3. **AWS CDK** — rejected; adds Node/TS or Go CDK toolchain.
4. **ClickOps (console only)** — rejected; not reproducible, violates IaC spirit.

## Tradeoffs

Document in README: **do not edit Lambda/ECR/IAM in console** without backporting script changes. Re-evaluate Terraform when adding prod HA or multi-region (M8).
