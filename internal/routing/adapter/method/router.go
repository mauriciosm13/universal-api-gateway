package method

import (
	"context"
	"fmt"
	"strings"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

type routeEntry struct {
	method    string
	upstreams []string
}

// Router resolves requests by HTTP method.
type Router struct {
	routes []routeEntry
}

// NewRouter returns a method router for the given routes.
func NewRouter(routes []config.MethodRoute) (*Router, error) {
	if len(routes) == 0 {
		return nil, fmt.Errorf("method router: at least one method route is required")
	}

	entries := make([]routeEntry, len(routes))
	for i, route := range routes {
		entries[i] = routeEntry{
			method:    strings.ToUpper(route.Method),
			upstreams: route.Upstreams,
		}
	}

	return &Router{routes: entries}, nil
}

// Resolve returns the first matching method route.
func (r *Router) Resolve(_ context.Context, req domain.Request) (port.Route, error) {
	method := strings.ToUpper(req.Method)

	for _, route := range r.routes {
		if method == route.method {
			return port.Route{
				ID:        fmt.Sprintf("method:%s", route.method),
				Upstreams: route.upstreams,
			}, nil
		}
	}

	return port.Route{}, port.ErrNoRoute
}
