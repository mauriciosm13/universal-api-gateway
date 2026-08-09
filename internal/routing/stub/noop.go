package stub

import (
	"context"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

// NoOpRouter is a scaffold router that never matches a route.
type NoOpRouter struct{}

// NewNoOpRouter returns a router stub for M0.
func NewNoOpRouter() *NoOpRouter {
	return &NoOpRouter{}
}

// Resolve always returns ErrNoRoute.
func (r *NoOpRouter) Resolve(_ context.Context, _ domain.Request) (port.Route, error) {
	return port.Route{}, port.ErrNoRoute
}
