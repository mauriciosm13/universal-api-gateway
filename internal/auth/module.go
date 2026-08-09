package auth

import (
	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
)

// Module registers auth port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	b.ProvideAuthenticator(authstub.NewNoOpAuthenticator())
}
