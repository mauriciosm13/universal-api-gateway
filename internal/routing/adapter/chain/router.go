package chain

import (
	"context"
	"errors"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

// Router tries each delegate in order until one resolves a route.
type Router struct {
	routers []port.Router
}

// NewRouter returns a chain of routers evaluated in order.
func NewRouter(routers ...port.Router) *Router {
	copied := make([]port.Router, len(routers))
	copy(copied, routers)

	return &Router{routers: copied}
}

// Resolve returns the first route resolved by a delegate router.
func (r *Router) Resolve(ctx context.Context, req domain.Request) (port.Route, error) {
	for _, router := range r.routers {
		route, err := router.Resolve(ctx, req)
		if err == nil {
			return route, nil
		}
		if !errors.Is(err, port.ErrNoRoute) {
			return port.Route{}, err
		}
	}

	return port.Route{}, port.ErrNoRoute
}
