package di

import (
	"context"
	"testing"

	authport "github.com/mauriciomendonca/universal-api-gateway/internal/auth/port"
	authstub "github.com/mauriciomendonca/universal-api-gateway/internal/auth/stub"
	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	middlewarestub "github.com/mauriciomendonca/universal-api-gateway/internal/middleware/stub"
	observabilitystub "github.com/mauriciomendonca/universal-api-gateway/internal/observability/stub"
	ratelimitstub "github.com/mauriciomendonca/universal-api-gateway/internal/ratelimit/stub"
	routingstub "github.com/mauriciomendonca/universal-api-gateway/internal/routing/stub"
)

func TestBuildWiresAllPorts(t *testing.T) {
	t.Parallel()

	deps, err := NewBuilder(config.Config{Host: "127.0.0.1", Port: 8080}, defaultTestModules()...).Build()
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

func TestBuildMissingProviderReturnsError(t *testing.T) {
	t.Parallel()

	_, err := NewBuilder(config.Config{}).Build()
	if err == nil {
		t.Fatal("expected error for incomplete graph")
	}
}

func TestBuilderAccessors(t *testing.T) {
	t.Parallel()

	custom := authstub.NewNoOpAuthenticator()
	builder := NewBuilder(config.Config{Host: "10.0.0.1"}, FuncModule{
		Name: "partial",
		Fn: func(b *Builder) {
			b.ProvideAuthenticator(custom)
			b.ProvideLimiter(ratelimitstub.NewNoOpLimiter())
		},
	})

	if builder.Config().Host != "10.0.0.1" {
		t.Fatalf("Config().Host = %q", builder.Config().Host)
	}
	if builder.Authenticator() != custom {
		t.Fatal("expected custom authenticator")
	}
	if builder.Limiter() == nil {
		t.Fatal("expected limiter")
	}
}

func TestModuleOverrideReplacesProvider(t *testing.T) {
	t.Parallel()

	custom := &stubAuthenticator{subject: "override"}

	modules := append(defaultTestModules(), FuncModule{
		Name: "override-auth",
		Fn: func(b *Builder) {
			b.ProvideAuthenticator(custom)
		},
	})

	deps, err := NewBuilder(config.Config{}, modules...).Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	identity, err := deps.Authenticator.AuthenticateRequest(context.Background(), domain.Request{})
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if identity.Subject != "override" {
		t.Fatalf("expected override authenticator, got subject %q", identity.Subject)
	}
}

type stubAuthenticator struct {
	subject string
}

func (s *stubAuthenticator) AuthenticateRequest(_ context.Context, _ domain.Request) (authport.Identity, error) {
	return authport.Identity{Subject: s.subject}, nil
}

func defaultTestModules() []Module {
	return []Module{
		FuncModule{Name: "auth", Fn: func(b *Builder) {
			b.ProvideAuthenticator(authstub.NewNoOpAuthenticator())
		}},
		FuncModule{Name: "routing", Fn: func(b *Builder) {
			b.ProvideRouter(routingstub.NewNoOpRouter())
		}},
		FuncModule{Name: "ratelimit", Fn: func(b *Builder) {
			b.ProvideLimiter(ratelimitstub.NewNoOpLimiter())
		}},
		FuncModule{Name: "middleware", Fn: func(b *Builder) {
			b.ProvidePipeline(middlewarestub.NewPassthroughPipeline())
		}},
		FuncModule{Name: "observability", Fn: func(b *Builder) {
			b.ProvideLogger(observabilitystub.NewNoOpLogger())
			b.ProvideTracer(observabilitystub.NewNoOpTracer())
		}},
	}
}
