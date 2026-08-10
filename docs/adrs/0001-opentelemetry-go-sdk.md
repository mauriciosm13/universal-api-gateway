# ADR 0001 — OpenTelemetry Go SDK

## Context

Milestone 0 requires OpenTelemetry setup aligned with the Project Bible principle **OpenTelemetry First**. The gateway already defines `observability/port.Tracer` with a no-op stub. Production-grade tracing needs a standard SDK and exporters without cloud-vendor lock-in.

## Decision

Adopt the official **OpenTelemetry Go SDK** (`go.opentelemetry.io/otel` and related modules) for trace bootstrap and the `port.Tracer` adapter.

Initial scope (M0):

- TracerProvider setup with stdout or OTLP/HTTP exporter
- Disabled by default via `OTEL_SDK_DISABLED=true` (no behavior change for existing deployments)
- Graceful shutdown of the TracerProvider on process exit
- Logger port remains no-op until M4 structured logging integration

Dependencies:

- `go.opentelemetry.io/otel`
- `go.opentelemetry.io/otel/sdk`
- `go.opentelemetry.io/otel/exporters/stdout/stdouttrace`
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`

Configuration uses OpenTelemetry environment variables (`OTEL_SERVICE_NAME`, `OTEL_TRACES_EXPORTER`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SDK_DISABLED`).

## Consequences

**Positive**

- Industry-standard tracing, cloud-agnostic exporters
- Enables M4 metrics/logs on the same OTel foundation
- Adapter pattern keeps domain and ports free of OTel imports

**Negative**

- New external dependencies (require module updates and CI cache refresh)
- Operators must configure env vars to enable tracing

## Alternatives

1. **Custom tracing interface only** — rejected; reinventing propagation and export.
2. **Vendor SDK (Datadog, New Relic)** — rejected; violates cloud-agnostic rule.
3. **Jaeger client directly** — rejected; OTel is the project standard and supersedes direct Jaeger clients.

## Tradeoffs

Stdout exporter is acceptable for M0 development; production uses OTLP to any compatible backend (Jaeger, Tempo, Honeycomb, cloud collectors) without code changes.
