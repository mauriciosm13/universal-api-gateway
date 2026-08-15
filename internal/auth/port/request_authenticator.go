package port

import (
	"context"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

// RequestAuthenticator validates credentials extracted from an inbound request.
type RequestAuthenticator interface {
	AuthenticateRequest(ctx context.Context, req domain.Request) (Identity, error)
}
