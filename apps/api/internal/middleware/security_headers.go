package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders sets standard defense-in-depth response headers on every
// request. X-Content-Type-Options stops browsers from sniffing a declared
// media type, Strict-Transport-Security pins future visits to HTTPS, and
// Referrer-Policy keeps request paths out of cross-origin Referer headers.
//
// Responses on key-management and admin routes additionally carry
// Cache-Control: no-store so browsers and shared caches never retain API keys
// or account data.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		h.Set("Referrer-Policy", "no-referrer")
		if noStorePath(r.URL.Path) {
			h.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

// noStorePrefixes lists the route prefixes whose responses must never be
// cached. Matches are exact or whole-segment; lookalike paths such as
// /api/v1/api-keys-public are not affected.
var noStorePrefixes = []string{
	"/api/v1/api-keys",
	"/api/v2/api-keys",
	"/api/v1/admin",
	"/api/v2/admin",
}

// noStorePath reports whether path belongs to a route carrying secret or
// account data.
func noStorePath(path string) bool {
	for _, prefix := range noStorePrefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}
