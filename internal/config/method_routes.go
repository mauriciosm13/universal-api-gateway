package config

import (
	"fmt"
	"net/http"
	"strings"
)

// MethodRoute maps an HTTP method to one or more upstream URLs.
type MethodRoute struct {
	Method    string
	Upstreams []string
}

// ParseMethodRoutes parses comma-separated METHOD=upstream pairs.
// Each upstream side may list multiple comma-separated URLs.
func ParseMethodRoutes(raw string) ([]MethodRoute, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	entries := splitRouteEntries(raw, isMethodRouteEntry)
	routes := make([]MethodRoute, 0, len(entries))

	for _, entry := range entries {
		method, upstreamRaw, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected METHOD=upstream", entry)
		}

		method = strings.TrimSpace(strings.ToUpper(method))
		if method == "" {
			return nil, fmt.Errorf("entry %q: method is required", entry)
		}
		if !isValidHTTPMethod(method) {
			return nil, fmt.Errorf("entry %q: invalid HTTP method %q", entry, method)
		}

		upstreams, err := ParseUpstreamList(upstreamRaw)
		if err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry, err)
		}

		routes = append(routes, MethodRoute{
			Method:    method,
			Upstreams: upstreams,
		})
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("GATEWAY_METHOD_ROUTES contains no valid entries")
	}

	return routes, nil
}

func isValidHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}
