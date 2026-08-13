package request

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
	"github.com/mauriciomendonca/universal-api-gateway/internal/middleware/port"
)

type errorBody struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorMiddleware converts handler errors into JSON domain responses.
type ErrorMiddleware struct{}

// NewErrorMiddleware returns middleware that maps errors to JSON responses.
func NewErrorMiddleware() *ErrorMiddleware {
	return &ErrorMiddleware{}
}

// Wrap implements port.Middleware.
func (m *ErrorMiddleware) Wrap(next port.Handler) port.Handler {
	return func(ctx context.Context, req domain.Request) (domain.Response, error) {
		resp, err := next(ctx, req)
		if err != nil {
			return jsonResponse(http.StatusInternalServerError, "middleware error"), nil
		}

		return resp, nil
	}
}

func jsonResponse(status int, message string) domain.Response {
	body, _ := json.Marshal(errorBody{
		Error:   http.StatusText(status),
		Code:    status,
		Message: message,
	})

	return domain.Response{
		StatusCode: status,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       body,
	}
}
