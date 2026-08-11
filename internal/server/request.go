package server

import (
	"net/http"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

func toDomainRequest(r *http.Request) domain.Request {
	headers := make(map[string][]string, len(r.Header))
	for key, values := range r.Header {
		copied := make([]string, len(values))
		copy(copied, values)
		headers[key] = copied
	}

	return domain.Request{
		Method:  r.Method,
		Path:    r.URL.Path,
		Host:    r.Host,
		Headers: headers,
	}
}
