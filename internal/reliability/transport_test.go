package reliability_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
	"github.com/mauriciomendonca/universal-api-gateway/internal/reliability"
)

func TestNewRoundTripperWithoutRetry(t *testing.T) {
	t.Parallel()

	rt := reliability.NewRoundTripper(config.ReliabilityConfig{
		UpstreamTimeout: time.Second,
		RetryMax:        0,
	})

	if rt == nil {
		t.Fatal("expected transport")
	}

	_, ok := rt.(http.RoundTripper)
	if !ok {
		t.Fatalf("type = %T, want http.RoundTripper", rt)
	}
}

func TestNewRoundTripperWithRetry(t *testing.T) {
	t.Parallel()

	rt := reliability.NewRoundTripper(config.ReliabilityConfig{
		UpstreamTimeout: 500 * time.Millisecond,
		RetryMax:        2,
		RetryBackoff:    10 * time.Millisecond,
	})

	if rt == nil {
		t.Fatal("expected transport")
	}
}
