package stub

import (
	"context"
	"testing"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

func TestPassthroughPipelineReturnsNotImplemented(t *testing.T) {
	t.Parallel()

	pipeline := NewPassthroughPipeline()
	_, resp, err := pipeline.Execute(context.Background(), domain.Request{Path: "/api"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != 501 {
		t.Fatalf("expected status 501, got %d", resp.StatusCode)
	}
}
