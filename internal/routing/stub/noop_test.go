package stub

import (
	"context"
	"errors"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

func TestNoOpRouterReturnsErrNoRoute(t *testing.T) {
	t.Parallel()

	router := NewNoOpRouter()
	_, err := router.Resolve(context.Background(), domain.Request{Path: "/unknown"})
	if !errors.Is(err, port.ErrNoRoute) {
		t.Fatalf("expected ErrNoRoute, got %v", err)
	}
}
