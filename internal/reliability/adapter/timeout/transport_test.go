package timeout_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mauriciomendonca/universal-api-gateway/internal/reliability/adapter/timeout"
)

type slowTransport struct {
	delay time.Duration
}

func (s slowTransport) RoundTrip(*http.Request) (*http.Response, error) {
	time.Sleep(s.delay)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(http.NoBody),
	}, nil
}

func TestRoundTripperEnforcesTimeout(t *testing.T) {
	t.Parallel()

	rt := timeout.NewRoundTripper(slowTransport{delay: 200 * time.Millisecond}, 50*time.Millisecond)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	_, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRoundTripperPassesFastResponses(t *testing.T) {
	t.Parallel()

	rt := timeout.NewRoundTripper(slowTransport{delay: 10 * time.Millisecond}, time.Second)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRoundTripperUsesDefaultBase(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	rt := timeout.NewRoundTripper(nil, time.Second)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	resp.Body.Close()
}
