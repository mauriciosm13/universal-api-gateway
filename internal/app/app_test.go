package app

import (
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
)

func TestNewBuildsDependencies(t *testing.T) {
	t.Parallel()

	deps := New(config.Config{Host: "127.0.0.1", Port: 8080})

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
