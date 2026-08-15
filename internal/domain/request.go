package domain

// Request is a transport-agnostic inbound gateway request.
type Request struct {
	Method  string
	Path    string
	Host    string
	Headers map[string][]string
	Query   map[string][]string
}

// Response is a transport-agnostic outbound gateway response.
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}
