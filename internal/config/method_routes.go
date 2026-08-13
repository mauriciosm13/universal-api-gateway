package config

import (
	"fmt"
	"net/http"
	"strings"
)

// MethodRoute maps an HTTP method to an upstream URL.
type MethodRoute struct {
	Method   string
	Upstream string
}

// ParseMethodRoutes parses comma-separated METHOD=upstream pairs.
func ParseMethodRoutes(raw string) ([]MethodRoute, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	entries := strings.Split(raw, ",")
	routes := make([]MethodRoute, 0, len(entries))

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		method, upstream, ok := strings.Cut(entry, "=")
		if !ok {
			return nil, fmt.Errorf("entry %q: expected METHOD=upstream", entry)
		}

		method = strings.TrimSpace(strings.ToUpper(method))
		upstream = strings.TrimSpace(upstream)
		if method == "" {
			return nil, fmt.Errorf("entry %q: method is required", entry)
		}
		if upstream == "" {
			return nil, fmt.Errorf("entry %q: upstream is required", entry)
		}
		if !isValidHTTPMethod(method) {
			return nil, fmt.Errorf("entry %q: invalid HTTP method %q", entry, method)
		}
		if err := ValidateUpstreamURL(upstream); err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry, err)
		}

		routes = append(routes, MethodRoute{
			Method:   method,
			Upstream: upstream,
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
