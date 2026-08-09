package port

import "context"

// Identity represents an authenticated caller.
type Identity struct {
	Subject string
	Claims  map[string]string
}

// Authenticator validates credentials and returns caller identity.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (Identity, error)
}
