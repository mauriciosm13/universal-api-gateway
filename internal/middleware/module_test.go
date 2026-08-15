package middleware

import (
	"testing"

	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	observabilitystub "github.com/mauriciomendonca/universal-api-gateway/internal/observability/stub"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

func TestModuleRegisterProvidesPipeline(t *testing.T) {
	t.Parallel()

	deps, err := di.NewBuilder(config.Config{},
		di.FuncModule{Name: "ports", Fn: func(b *di.Builder) {
			b.ProvideAuthenticator(authstub.NewNoOpAuthenticator())
			b.ProvideRouter(routingstub.NewNoOpRouter())
			b.ProvideLimiter(ratelimitstub.NewNoOpLimiter())
			b.ProvideLogger(observabilitystub.NewNoOpLogger())
			b.ProvideTracer(observabilitystub.NewNoOpTracer())
		}},
		Module{},
	).Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if deps.Pipeline == nil {
		t.Fatal("expected pipeline from middleware module")
	}
}
