package observability

import (
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	"github.com/mauriciomendonca/universal-api-gateway/internal/observability/adapter"
	observabilitystub "github.com/mauriciomendonca/universal-api-gateway/internal/observability/stub"
)

// Module registers observability port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	if b.Config().Telemetry.Disabled {
		b.ProvideLogger(observabilitystub.NewNoOpLogger())
		b.ProvideTracer(observabilitystub.NewNoOpTracer())
		return
	}

	b.ProvideLogger(observabilitystub.NewNoOpLogger())
	b.ProvideTracer(adapter.NewTracer())
}
