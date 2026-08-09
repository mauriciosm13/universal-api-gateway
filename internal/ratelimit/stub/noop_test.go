package stub

import (
	"context"
	"testing"
)

func TestNoOpLimiterAllowsAll(t *testing.T) {
	t.Parallel()

	limiter := NewNoOpLimiter()
	allowed, err := limiter.Allow(context.Background(), "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !allowed {
		t.Fatal("expected request to be allowed")
	}
}
