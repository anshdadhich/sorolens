package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// securityHeadersRequest runs SecurityHeaders with one request and returns the
// recorded response.
func securityHeadersRequest(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	SecurityHeaders(next).ServeHTTP(rec, req)
	return rec
}

func TestSecurityHeadersAreSet(t *testing.T) {
	t.Parallel()

	rec := securityHeadersRequest(t, "/api/v1/contracts")
	want := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "no-referrer",
	}
	for header, value := range want {
		if got := rec.Header().Get(header); got != value {
			t.Errorf("%s = %q, want %q", header, got, value)
		}
	}
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Errorf("Cache-Control = %q, want empty on a cacheable route", got)
	}
}

func TestSecurityHeadersNoStoreOnSensitiveRoutes(t *testing.T) {
	t.Parallel()

	paths := []string{
		"/api/v1/api-keys",
		"/api/v1/api-keys/key_123",
		"/api/v2/api-keys",
		"/api/v2/api-keys/key_123/rotate",
		"/api/v1/admin",
		"/api/v1/admin/users",
		"/api/v2/admin",
		"/api/v2/admin/stats",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			rec := securityHeadersRequest(t, path)
			if got := rec.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store for %s", got, path)
			}
		})
	}
}

func TestSecurityHeadersSkipsCacheControlOnLookalikes(t *testing.T) {
	t.Parallel()

	paths := []string{
		"/api/v1/api-keys-public",
		"/api/v1/admin-panel",
		"/api/v1/contracts",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			rec := securityHeadersRequest(t, path)
			if got := rec.Header().Get("Cache-Control"); got != "" {
				t.Errorf("Cache-Control = %q, want empty for %s", got, path)
			}
		})
	}
}
