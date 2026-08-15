package credential

import (
	"strings"

	"github.com/mauriciomendonca/universal-api-gateway/internal/domain"
)

// HasAuthorizationHeader reports whether Authorization header is present.
func HasAuthorizationHeader(req domain.Request) bool {
	return HeaderValue(req, "Authorization") != ""
}

// BearerToken extracts token from Authorization: Bearer <token>.
// Malformed or missing Bearer prefix returns empty string.
func BearerToken(req domain.Request) string {
	value := HeaderValue(req, "Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(value, prefix))
}

// APIKey returns the API key from header or query, header first.
func APIKey(req domain.Request, headerName, queryParam string) string {
	if key := HeaderValue(req, headerName); key != "" {
		return key
	}

	return QueryValue(req, queryParam)
}

// HeaderValue returns the first value for a case-insensitive header name.
func HeaderValue(req domain.Request, name string) string {
	for key, values := range req.Headers {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
	}

	return ""
}

// QueryValue returns the first value for an exact query param name.
func QueryValue(req domain.Request, name string) string {
	for key, values := range req.Query {
		if key == name && len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
	}

	return ""
}
