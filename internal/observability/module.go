package observability

import (
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	observabilitystub "github.com/mauriciomendonca/universal-api-gateway/internal/observability/stub"
)

// Module registers observability port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	b.ProvideLogger(observabilitystub.NewNoOpLogger())
	b.ProvideTracer(observabilitystub.NewNoOpTracer())
}
