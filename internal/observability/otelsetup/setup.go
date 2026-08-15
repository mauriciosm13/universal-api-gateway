package otelsetup

import (
	"context"
	"fmt"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Install configures the global OpenTelemetry TracerProvider when telemetry is enabled.
func Install(ctx context.Context, cfg config.TelemetryConfig) (func(context.Context) error, error) {
	noopShutdown := func(context.Context) error { return nil }
	if cfg.Disabled {
		return noopShutdown, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	var provider *sdktrace.TracerProvider

	switch cfg.TracesExporter {
	case "none":
		provider = sdktrace.NewTracerProvider(sdktrace.WithResource(res))
	case "stdout", "otlp":
		exporter, err := newExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
		provider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
		)
	default:
		return nil, fmt.Errorf("unsupported traces exporter %q", cfg.TracesExporter)
	}

	otel.SetTracerProvider(provider)

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return provider.Shutdown(ctx)
	}, nil
}

func newExporter(ctx context.Context, cfg config.TelemetryConfig) (sdktrace.SpanExporter, error) {
	switch cfg.TracesExporter {
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "otlp":
		if cfg.OTLPEndpoint == "" {
			return nil, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT required when OTEL_TRACES_EXPORTER=otlp")
		}
		return otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(cfg.OTLPEndpoint))
	default:
		return nil, fmt.Errorf("unsupported traces exporter %q", cfg.TracesExporter)
	}
}
