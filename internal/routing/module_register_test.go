package routing

import (
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
)

func TestModuleRegisterPanicsOnInvalidConfig(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Register() did not panic on invalid router config")
		}
	}()

	di.NewBuilder(config.Config{DefaultUpstream: "not-a-url"}, Module{})
}
