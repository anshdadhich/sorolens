// Package middleware provides shared net/http middleware for the Sorolens
// API server.
package middleware

import (
	"mime"
	"net/http"
	"strings"
)

// RequireJSONContentType returns middleware that rejects POST and PUT
// requests whose Content-Type header is missing or is not
// "application/json". Matching requests are passed through to next
// unchanged; all other methods are never inspected.
//
// A rejected request receives a 415 Unsupported Media Type response in the
// standard JSON error envelope with code UNSUPPORTED_MEDIA_TYPE.
func RequireJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			WriteError(w, r, http.StatusUnsupportedMediaType, CodeUnsupportedMedia, "Content-Type must be application/json")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ContentTypeJSON rejects POST and PUT requests whose Content-Type is not application/json.
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			ct := r.Header.Get("Content-Type")
			if !strings.HasPrefix(ct, "application/json") {
				WriteError(w, r, http.StatusUnsupportedMediaType, CodeUnsupportedMedia, "Content-Type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
