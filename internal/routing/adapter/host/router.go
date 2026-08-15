package host

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/routing/port"
)

type routeEntry struct {
	host      string
	upstreams []string
}

// Router resolves requests by HTTP Host header.
type Router struct {
	routes []routeEntry
}

// NewRouter returns a host router for the given routes.
func NewRouter(routes []config.HostRoute) (*Router, error) {
	if len(routes) == 0 {
		return nil, fmt.Errorf("host router: at least one host route is required")
	}

	entries := make([]routeEntry, len(routes))
	for i, route := range routes {
		entries[i] = routeEntry{
			host:      strings.ToLower(route.Host),
			upstreams: route.Upstreams,
		}
	}

	return &Router{routes: entries}, nil
}

// Resolve returns the first matching host route.
func (r *Router) Resolve(_ context.Context, req domain.Request) (port.Route, error) {
	requestHost := normalizeHost(req.Host)

	for _, route := range r.routes {
		if requestHost == route.host {
			return port.Route{
				ID:        fmt.Sprintf("host:%s", route.host),
				Upstreams: route.upstreams,
			}, nil
		}
	}

	return port.Route{}, port.ErrNoRoute
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return host
	}

	if name, _, err := net.SplitHostPort(host); err == nil {
		return name
	}

	return host
}
