package server

import (
	"context"
	"net"
	"net/http"
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

func TestStartAndShutdown(t *testing.T) {
	t.Parallel()

	deps := Dependencies{
		Config: config.Config{
			Host:         "127.0.0.1",
			Port:         0,
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
	}

	srv := New(deps)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	srv.httpServer.Addr = ln.Addr().String()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.httpServer.Serve(ln)
	}()

	resp, err := http.Get("http://" + ln.Addr().String() + "/health")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	if err := <-errCh; err != http.ErrServerClosed {
		t.Fatalf("Serve() error = %v", err)
	}
}

func TestStartInvalidAddress(t *testing.T) {
	t.Parallel()

	deps := Dependencies{
		Config: config.Config{
			Host:         "invalid host",
			Port:         8080,
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
			IdleTimeout:  time.Second,
		},
	}

	srv := New(deps)
	if err := srv.Start(); err == nil {
		t.Fatal("expected Start() error for invalid listen address")
	}
}
