package server

import (
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
)

func TestNewRegistersHealthServer(t *testing.T) {
	t.Parallel()

	deps := Dependencies{
		Config: config.Config{
			Host:         "127.0.0.1",
			Port:         18080,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	srv := New(deps)
	if srv == nil {
		t.Fatal("expected non-nil server")
	}

	if srv.httpServer == nil {
		t.Fatal("expected non-nil http server")
	}

	if srv.httpServer.Addr != "127.0.0.1:18080" {
		t.Fatalf("expected addr 127.0.0.1:18080, got %q", srv.httpServer.Addr)
	}
}
