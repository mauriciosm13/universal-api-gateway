package otelsetup

import (
	"context"
	"strings"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
)

func TestInstallDisabled(t *testing.T) {
	t.Parallel()

	shutdown, err := Install(context.Background(), config.TelemetryConfig{Disabled: true})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown func")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown() error = %v", err)
	}
}

func TestInstallUnsupportedExporter(t *testing.T) {
	t.Parallel()

	_, err := Install(context.Background(), config.TelemetryConfig{
		ServiceName:    "gw",
		TracesExporter: "bogus",
	})
	if err == nil {
		t.Fatal("Install() error = nil, want unsupported exporter")
	}
	if !strings.Contains(err.Error(), "unsupported traces exporter") {
		t.Fatalf("Install() error = %v, want unsupported traces exporter", err)
	}
}

func TestInstallOTLPMissingEndpoint(t *testing.T) {
	t.Parallel()

	_, err := Install(context.Background(), config.TelemetryConfig{
		ServiceName:    "gw",
		TracesExporter: "otlp",
	})
	if err == nil {
		t.Fatal("Install() error = nil, want missing OTLP endpoint")
	}
	if !strings.Contains(err.Error(), "OTEL_EXPORTER_OTLP_ENDPOINT") {
		t.Fatalf("Install() error = %v, want OTLP endpoint message", err)
	}
}

func TestInstallNoneExporter(t *testing.T) {
	shutdown, err := Install(context.Background(), config.TelemetryConfig{
		ServiceName:    "gw",
		TracesExporter: "none",
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	t.Cleanup(func() {
		_ = shutdown(context.Background())
	})
}
