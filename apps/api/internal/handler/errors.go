package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sorolens/sorolens/apps/api/internal/middleware"
)

// Error codes live in middleware so every error response, from middleware or
// handler, draws on one vocabulary (issue #533).
const (
	CodeNotFound         = middleware.CodeNotFound
	CodeInvalidInput     = middleware.CodeInvalidInput
	CodeInternal         = middleware.CodeInternal
	CodeForbidden        = middleware.CodeForbidden
	CodeRateLimited      = middleware.CodeRateLimited
	CodeUnsupportedMedia = middleware.CodeUnsupportedMedia
	CodeUnauthorized     = middleware.CodeUnauthorized
)

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	middleware.WriteError(w, r, status, code, message)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
