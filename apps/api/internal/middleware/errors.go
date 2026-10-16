package middleware

import (
	"encoding/json"
	"net/http"
)

// Error codes shared by middleware and handlers, so every failure carries a
// machine-readable code from one vocabulary (issue #533).
const (
	CodeNotFound         = "NOT_FOUND"
	CodeInvalidInput     = "INVALID_INPUT"
	CodeInternal         = "INTERNAL"
	CodeForbidden        = "FORBIDDEN"
	CodeRateLimited      = "RATE_LIMITED"
	CodeUnsupportedMedia = "UNSUPPORTED_MEDIA_TYPE"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodePayloadTooLarge  = "PAYLOAD_TOO_LARGE"
	CodeTimeout          = "TIMEOUT"
)

// ErrorBody is the payload inside every JSON error response.
type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// ErrorEnvelope is the one shape every error response uses: the details under
// an "error" key, so SDKs and the web app parse a single structure.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// WriteError writes the standard JSON error envelope, stamping the request ID
// from r's context so a failure can be traced in the logs. Middleware and
// handlers both route through here; handler.writeError delegates to it.
func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope(r, code, message))
}

// ErrorJSON renders the standard envelope as a string for callers that need
// the body before a ResponseWriter exists (http.TimeoutHandler).
func ErrorJSON(r *http.Request, code, message string) string {
	b, _ := json.Marshal(envelope(r, code, message))
	return string(b)
}

func envelope(r *http.Request, code, message string) ErrorEnvelope {
	return ErrorEnvelope{
		Error: ErrorBody{
			Code:      code,
			Message:   message,
			RequestID: GetRequestID(r.Context()),
		},
	}
}
