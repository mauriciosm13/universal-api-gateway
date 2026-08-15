package observability

import (
	"testing"

	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	middlewarestub "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/observability/adapter"
	observabilitystub "github.com/mauriciomendonca/universal-api-gateway/internal/observability/stub"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

func TestModuleRegisterDisabledUsesNoOpTracer(t *testing.T) {
	t.Parallel()

	deps, err := di.NewBuilder(config.Config{Telemetry: config.TelemetryConfig{Disabled: true}},
		scaffoldPorts(),
		Module{},
	).Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if _, ok := deps.Tracer.(*observabilitystub.NoOpTracer); !ok {
		t.Fatalf("tracer type = %T, want *stub.NoOpTracer", deps.Tracer)
	}
}

func TestModuleRegisterEnabledUsesAdapterTracer(t *testing.T) {
	t.Parallel()

	deps, err := di.NewBuilder(config.Config{},
		scaffoldPorts(),
		Module{},
	).Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if _, ok := deps.Tracer.(*adapter.Tracer); !ok {
		t.Fatalf("tracer type = %T, want *adapter.Tracer", deps.Tracer)
	}
}

func scaffoldPorts() di.Module {
	return di.FuncModule{Name: "ports", Fn: func(b *di.Builder) {
		b.ProvideAuthenticator(authstub.NewNoOpAuthenticator())
		b.ProvideRouter(routingstub.NewNoOpRouter())
		b.ProvideLimiter(ratelimitstub.NewNoOpLimiter())
		b.ProvidePipeline(middlewarestub.NewPassthroughPipeline())
	}}
}
