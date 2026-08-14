package auth

import (
	"fmt"

	jwtadapter "github.com/mauriciomendonca/universal-api-gateway/internal/auth/adapter/jwt"
	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
)

// Module registers auth port providers for the dependency graph.
type Module struct{}

// Register implements di.Module.
func (Module) Register(b *di.Builder) {
	authenticator, err := buildAuthenticator(b.Config())
	if err != nil {
		panic(fmt.Sprintf("auth: %v", err))
	}

	b.ProvideAuthenticator(authenticator)
}

func buildAuthenticator(cfg config.Config) (authport.Authenticator, error) {
	if !cfg.JWT.Enabled() {
		return authstub.NewNoOpAuthenticator(), nil
	}

	return jwtadapter.NewValidator(cfg.JWT)
}
