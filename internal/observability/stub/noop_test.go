package stub

import (
	"context"
	"testing"
)

func TestNoOpLoggerAndTracer(t *testing.T) {
	t.Parallel()

	logger := NewNoOpLogger()
	logger.Info("ignored")
	logger.Error("ignored")

	tracer := NewNoOpTracer()
	ctx, finish := tracer.StartSpan(context.Background(), "test")
	finish()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
}
