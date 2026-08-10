package adapter

import (
	"context"

	"github.com/mauriciomendonca/universal-api-gateway/internal/observability/port"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/mauriciomendonca/universal-api-gateway"

// Tracer implements port.Tracer using the global OpenTelemetry TracerProvider.
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer returns a tracer backed by the installed OpenTelemetry provider.
func NewTracer() *Tracer {
	return &Tracer{tracer: otel.Tracer(instrumentationName)}
}

// StartSpan creates an OpenTelemetry span and returns a finish function.
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, func()) {
	ctx, span := t.tracer.Start(ctx, name)
	return ctx, func() { span.End() }
}

var _ port.Tracer = (*Tracer)(nil)
