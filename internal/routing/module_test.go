package routing

import (
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	chainrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/chain"
	hostrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/host"
	methodrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/method"
	staticrouter "github.com/mauriciomendonca/universal-api-gateway/internal/routing/adapter/static"
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

func TestBuildRouterStaticOnly(t *testing.T) {
	t.Parallel()

	router, err := buildRouter(config.Config{
		DefaultUpstream: "http://default:8080",
	})
	if err != nil {
		t.Fatalf("buildRouter() error = %v", err)
	}

	if _, ok := router.(*staticrouter.Router); !ok {
		t.Fatalf("router type = %T, want *static.Router", router)
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

	if _, ok := router.(*chainrouter.Router); !ok {
		t.Fatalf("router type = %T, want *chain.Router", router)
	}
}

func TestBuildRouterPathAndDefaultUsesChain(t *testing.T) {
	t.Parallel()

	router, err := buildRouter(config.Config{
		PathRoutes: []config.PathRoute{
			{Prefix: "/api", Upstream: "http://api:8080"},
		},
		DefaultUpstream: "http://default:8080",
	})
	if err != nil {
		t.Fatalf("buildRouter() error = %v", err)
	}

	if _, ok := router.(*chainrouter.Router); !ok {
		t.Fatalf("router type = %T, want *chain.Router", router)
	}
}

func TestBuildRouterHostAndMethodChain(t *testing.T) {
	t.Parallel()

	router, err := buildRouter(config.Config{
		HostRoutes: []config.HostRoute{
			{Host: "api.example.com", Upstream: "http://host:8080"},
		},
		MethodRoutes: []config.MethodRoute{
			{Method: "POST", Upstream: "http://post:8080"},
		},
	})
	if err != nil {
		t.Fatalf("buildRouter() error = %v", err)
	}

	if _, ok := router.(*chainrouter.Router); !ok {
		t.Fatalf("router type = %T, want *chain.Router", router)
	}
}

func TestBuildRouterInvalidDefaultUpstream(t *testing.T) {
	t.Parallel()

	_, err := buildRouter(config.Config{DefaultUpstream: "not-a-url"})
	if err == nil {
		t.Fatal("buildRouter() error = nil, want invalid default upstream")
	}
}

func TestBuildRouterSingleHostOnly(t *testing.T) {
	t.Parallel()

	router, err := buildRouter(config.Config{
		HostRoutes: []config.HostRoute{
			{Host: "api.example.com", Upstream: "http://host:8080"},
		},
	})
	if err != nil {
		t.Fatalf("buildRouter() error = %v", err)
	}

	if _, ok := router.(*hostrouter.Router); !ok {
		t.Fatalf("router type = %T, want *host.Router", router)
	}
}

func TestBuildRouterSingleMethodOnly(t *testing.T) {
	t.Parallel()

	router, err := buildRouter(config.Config{
		MethodRoutes: []config.MethodRoute{
			{Method: "POST", Upstream: "http://post:8080"},
		},
	})
	if err != nil {
		t.Fatalf("buildRouter() error = %v", err)
	}

	if _, ok := router.(*methodrouter.Router); !ok {
		t.Fatalf("router type = %T, want *method.Router", router)
	}
}
