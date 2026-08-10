package adapter

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTracerStartSpanRecordsSpan(t *testing.T) {
	t.Parallel()

	sr := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	tracer := provider.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	if spans[0].Name() != "test-span" {
		t.Fatalf("expected span name test-span, got %q", spans[0].Name())
	}
}

func TestPortTracerStartSpan(t *testing.T) {
	t.Parallel()

	sr := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	tr := &Tracer{tracer: provider.Tracer("test")}

	_, finish := tr.StartSpan(context.Background(), "gateway-span")
	finish()

	if len(sr.Ended()) != 1 {
		t.Fatalf("expected 1 span, got %d", len(sr.Ended()))
	}
}
