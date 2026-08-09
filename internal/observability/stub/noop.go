package stub

import (
	"context"

	"github.com/mauriciomendonca/universal-api-gateway/internal/observability/port"
)

// NoOpLogger discards structured log events.
type NoOpLogger struct{}

// NewNoOpLogger returns a scaffold logger for M0.
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// Info discards the log event.
func (l *NoOpLogger) Info(_ string, _ ...any) {}

// Error discards the log event.
func (l *NoOpLogger) Error(_ string, _ ...any) {}

// NoOpTracer discards span lifecycle events.
type NoOpTracer struct{}

// NewNoOpTracer returns a scaffold tracer for M0.
func NewNoOpTracer() *NoOpTracer {
	return &NoOpTracer{}
}

// StartSpan returns the input context and a no-op finish function.
func (t *NoOpTracer) StartSpan(ctx context.Context, _ string) (context.Context, func()) {
	return ctx, func() {}
}

var (
	_ port.Logger = (*NoOpLogger)(nil)
	_ port.Tracer = (*NoOpTracer)(nil)
)
