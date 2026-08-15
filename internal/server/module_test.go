package server

import (
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
)

func TestModuleRegisterIsNoOp(t *testing.T) {
	t.Parallel()

	builder := di.NewBuilder(config.Config{}, Module{})
	if builder.Config().Host != "" && builder.Config().Port != 0 {
		t.Fatalf("unexpected config mutation: %+v", builder.Config())
	}
}
