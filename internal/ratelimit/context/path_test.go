package context

import (
	"context"
	"testing"
)

func TestRequestPathRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := WithRequestPath(context.Background(), "/api/orders")
	got, ok := RequestPathFrom(ctx)
	if !ok {
		t.Fatal("RequestPathFrom() ok = false")
	}
	if got != "/api/orders" {
		t.Fatalf("path = %q, want /api/orders", got)
	}
}

func TestRequestPathFromEmpty(t *testing.T) {
	t.Parallel()

	if _, ok := RequestPathFrom(context.Background()); ok {
		t.Fatal("RequestPathFrom() ok = true, want false")
	}
}
