package stub

import (
	"context"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

// PassthroughPipeline is a scaffold pipeline for M0.
type PassthroughPipeline struct{}

// NewPassthroughPipeline returns a pipeline stub that responds not implemented.
func NewPassthroughPipeline() *PassthroughPipeline {
	return &PassthroughPipeline{}
}

// Execute returns a not-implemented response for non-health traffic.
func (p *PassthroughPipeline) Execute(_ context.Context, _ domain.Request) (domain.Response, error) {
	return domain.Response{
		StatusCode: 501,
		Headers:    map[string][]string{"Content-Type": {"text/plain"}},
		Body:       []byte("gateway pipeline not implemented"),
	}, nil
}
