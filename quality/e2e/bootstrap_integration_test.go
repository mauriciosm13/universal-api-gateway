//go:build integration

package e2e_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

type statusError struct {
	code int
}

func (e *statusError) Error() string {
	return fmt.Sprintf("status %d", e.code)
}

func demoBaseURL() string {
	if u := os.Getenv("GATEWAY_URL"); u != "" {
		return u
	}
	return "http://localhost:8080"
}

func TestDemoStackHealthReady(t *testing.T) {
	base := demoBaseURL()
	client := &http.Client{Timeout: 5 * time.Second}

	var lastErr error
	for range 60 {
		resp, err := client.Get(base + "/health/ready")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
			lastErr = &statusError{code: resp.StatusCode}
		} else {
			lastErr = err
		}
		time.Sleep(time.Second)
	}

	t.Fatalf("GET /health/ready not OK at %s: %v", base, lastErr)
}
