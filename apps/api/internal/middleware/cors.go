package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORS response values shared by every response. Preflight answers advertise
// the full method and header set the dashboard and API key clients use, and a
// ten-minute Access-Control-Max-Age lets the browser cache the preflight
// result.
const (
	corsMaxAge         = "600"
	corsAllowedHeaders = "Content-Type, X-Request-ID, X-User-ID, Authorization, X-API-Key"
	corsAllowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
)

// CORSFromEnv returns middleware that applies the CORS rules from
// CORS_ALLOWED_ORIGINS, the variable documented in the README: a
// comma-separated list of origins, or unset (or "*") to allow every origin.
// The variable is read once, when the middleware is built; restart the process
// to pick up a change.
func CORSFromEnv() func(http.Handler) http.Handler {
	return corsFromOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
}

// corsFromOrigins builds the CORS middleware for one raw CORS_ALLOWED_ORIGINS
// value.
//
// With no configured origins every response carries
// Access-Control-Allow-Origin: *. With a configured list, a request whose
// Origin header is in the list receives that origin echoed back; any other
// request falls back to the first configured origin (the browser still blocks
// the mismatch, but the header is always set). Every allowlist response is
// marked Vary: Origin so caches never confuse a permitted origin with a
// rejected one. Preflight (OPTIONS) requests are answered with 204 and the
// full header set before the router runs.
func corsFromOrigins(raw string) func(http.Handler) http.Handler {
	var allowed []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if o == "*" {
			allowed = nil
			break
		}
		allowed = append(allowed, o)
	}
	wildcard := len(allowed) == 0

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowOrigin := "*"
			if !wildcard {
				allowOrigin = allowed[0]
				for _, o := range allowed {
					if o == r.Header.Get("Origin") {
						allowOrigin = o
						break
					}
				}
				w.Header().Set("Vary", "Origin")
			}

			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Headers", corsAllowedHeaders)
			w.Header().Set("Access-Control-Allow-Methods", corsAllowedMethods)
			w.Header().Set("Access-Control-Max-Age", corsMaxAge)
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
