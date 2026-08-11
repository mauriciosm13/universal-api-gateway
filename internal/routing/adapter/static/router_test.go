package static

import (
	"context"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

func TestNewRouterValidUpstream(t *testing.T) {
	t.Parallel()

	router, err := NewRouter("http://backend:8080")
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	route, err := router.Resolve(context.Background(), domain.Request{Path: "/api/users"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if route.ID != "default" {
		t.Fatalf("expected route ID default, got %q", route.ID)
	}

	if route.Upstream != "http://backend:8080" {
		t.Fatalf("expected upstream http://backend:8080, got %q", route.Upstream)
	}
}

func TestNewRouterInvalidUpstream(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		upstream string
	}{
		{name: "missing scheme", upstream: "backend:8080"},
		{name: "missing host", upstream: "http://"},
		{name: "unsupported scheme", upstream: "ftp://backend:8080"},
		{name: "empty", upstream: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewRouter(tt.upstream)
			if err == nil {
				t.Fatal("expected error for invalid upstream")
			}
		})
	}
}

func TestResolveSameRouteForDifferentPaths(t *testing.T) {
	t.Parallel()

	router, err := NewRouter("https://upstream.example")
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	for _, path := range []string{"/", "/v1", "/nested/path"} {
		route, err := router.Resolve(context.Background(), domain.Request{Path: path})
		if err != nil {
			t.Fatalf("Resolve(%q) error = %v", path, err)
		}

		if route.Upstream != "https://upstream.example" {
			t.Fatalf("Resolve(%q) upstream = %q", path, route.Upstream)
		}
	}
}
