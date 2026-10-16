package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/sorolens/sorolens/apps/api/internal/store"
)

// ProtectMetricsFromEnv returns middleware that gates GET /metrics behind the
// optional METRICS_TOKEN environment variable. When the variable is unset or
// empty the endpoint stays open, preserving the default local setup; a
// deployment that exposes the API port publicly sets it so only scrapers and
// signed-in admins can read the exposition (issue #537).
func ProtectMetricsFromEnv(lookup UserLookup, logger *slog.Logger) func(http.Handler) http.Handler {
	return protectMetrics(strings.TrimSpace(os.Getenv("METRICS_TOKEN")), lookup, logger)
}

// protectMetrics is ProtectMetricsFromEnv with the token injected so tests can
// exercise it without touching the environment.
//
// A request passes when it presents the token as an Authorization: Bearer
// credential or resolves to a user holding the admin role (the same identity
// headers the role middleware reads). Everything else is rejected with 401 and
// a WWW-Authenticate challenge.
func protectMetrics(token string, lookup UserLookup, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if token == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if presented, ok := bearerCredential(r); ok &&
				subtle.ConstantTimeCompare([]byte(presented), []byte(token)) == 1 {
				next.ServeHTTP(w, r)
				return
			}
			if user, ok := resolveUser(r, lookup); ok && store.RoleRank(user.Role) >= store.RoleRank(store.RoleAdmin) {
				next.ServeHTTP(w, r)
				return
			}
			if logger != nil {
				logger.Debug("metrics middleware: rejected", "path", r.URL.Path)
			}
			w.Header().Set("WWW-Authenticate", "Bearer")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":       "UNAUTHORIZED",
					"message":    "authentication required",
					"request_id": GetRequestID(r.Context()),
				},
			})
		})
	}
}

// bearerCredential returns the token from an Authorization: Bearer header.
func bearerCredential(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", false
	}
	return strings.TrimSpace(h[len(prefix):]), true
}
