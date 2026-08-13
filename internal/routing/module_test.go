package routing

import (
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

func TestBuildRouterNoConfig(t *testing.T) {
	t.Parallel()

	router, err := buildRouter(config.Config{})
	if err != nil {
		t.Fatalf("buildRouter() error = %v", err)
	}

	if _, ok := router.(*routingstub.NoOpRouter); !ok {
		t.Fatalf("router type = %T, want *stub.NoOpRouter", router)
	}
}

func TestBuildRouterChain(t *testing.T) {
	t.Parallel()

	router, err := buildRouter(config.Config{
		HeaderRoutes: []config.HeaderRoute{
			{Name: "X-Version", Value: "v1", Upstream: "http://v1:8080"},
		},
		PathRoutes: []config.PathRoute{
			{Prefix: "/api", Upstream: "http://api:8080"},
		},
		DefaultUpstream: "http://default:8080",
	})
	if err != nil {
		t.Fatalf("buildRouter() error = %v", err)
	}

	if router == nil {
		t.Fatal("router = nil")
	}
}
