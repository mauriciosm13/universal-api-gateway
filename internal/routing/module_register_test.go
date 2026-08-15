package routing

import (
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
)

func TestModuleRegisterPanicsOnInvalidUpstream(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Register() did not panic on invalid upstream")
		}
	}()

	di.NewBuilder(config.Config{DefaultUpstream: "not-a-url"}, Module{})
}

func TestModuleRegisterProvidesRouterViaBuildRouter(t *testing.T) {
	t.Parallel()

	router, err := buildRouter(config.Config{DefaultUpstream: "http://default:8080"})
	if err != nil {
		t.Fatalf("buildRouter() error = %v", err)
	}
	if router == nil {
		t.Fatal("expected router")
	}
}
