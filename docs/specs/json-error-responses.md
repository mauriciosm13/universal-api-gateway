# Specification — JSON Error Responses (M1)

## Overview

Gateway-generated HTTP errors return a structured JSON body instead of plain text.

## Responsibilities

| Package | Responsibility |
|---|---|
| `internal/server/json_error.go` | Encode and write JSON error responses |
| `internal/server/gateway.go` | Use JSON errors for routing/upstream failures |

## API

```go
type ErrorBody struct {
    Error   string `json:"error"`
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func writeJSONError(w http.ResponseWriter, status int, message string)
```

## Response

- Status: 4xx or 5xx
- Header: `Content-Type: application/json`
- Body:

```json
{
  "error": "Not Found",
  "code": 404,
  "message": "no route matched"
}
```

`error` uses `http.StatusText(status)`.

## Scope

| Condition | JSON error |
|---|---|
| No route matched | Yes — 404 |
| Routing error | Yes — 500 |
| Invalid upstream in handler | Yes — 500 |
| Upstream unreachable (ReverseProxy) | No — stdlib behavior |
| Health endpoints | No — unchanged |

## Testing

- 404 response is JSON with correct fields and content type
- 500 routing error is JSON
- Existing proxy success paths unchanged
