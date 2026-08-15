package app

import (
	"context"
	"testing"

	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/di"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

func TestBuildWiresDefaultModules(t *testing.T) {
	t.Parallel()

	deps, err := Build(config.Config{Host: "127.0.0.1", Port: 8080})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if deps.Config.Host != "127.0.0.1" {
		t.Fatalf("expected host 127.0.0.1, got %q", deps.Config.Host)
	}

	if deps.Authenticator == nil {
		t.Fatal("expected non-nil authenticator")
	}
	if deps.Router == nil {
		t.Fatal("expected non-nil router")
	}
	if deps.Limiter == nil {
		t.Fatal("expected non-nil limiter")
	}
	if deps.Pipeline == nil {
		t.Fatal("expected non-nil pipeline")
	}
	if deps.Logger == nil {
		t.Fatal("expected non-nil logger")
	}
	if deps.Tracer == nil {
		t.Fatal("expected non-nil tracer")
	}
}

func TestBuildStubSwapViaModuleOverride(t *testing.T) {
	t.Parallel()

	custom := &stubAuthenticator{subject: "swapped"}

	modules := append(DefaultModules(), di.FuncModule{
		Name: "test-override",
		Fn: func(b *di.Builder) {
			b.ProvideAuthenticator(custom)
		},
	})

	deps, err := Build(config.Config{}, modules...)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	identity, err := deps.Authenticator.AuthenticateRequest(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if identity.Subject != "swapped" {
		t.Fatalf("expected swapped authenticator, got subject %q", identity.Subject)
	}
}

type stubAuthenticator struct {
	subject string
}

func (s *stubAuthenticator) AuthenticateRequest(_ context.Context, _ domain.Request) (authport.Identity, error) {
	return authport.Identity{Subject: s.subject}, nil
}
