package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// corsRequest runs mw with one request and reports whether the wrapped
// handler was reached.
func corsRequest(t *testing.T, mw func(http.Handler) http.Handler, method, origin string) (*httptest.ResponseRecorder, bool) {
	t.Helper()

	nextHit := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextHit = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(method, "/api/v1/contracts", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)
	return rec, nextHit
}

func TestCORSWildcardWhenUnset(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
	}{
		{"unset", ""},
		{"explicit wildcard", "*"},
		{"wildcard among origins", "https://one.example, *"},
		{"only whitespace", " , , "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec, nextHit := corsRequest(t, corsFromOrigins(tc.raw), http.MethodGet, "https://any.example")
			if !nextHit {
				t.Fatal("next handler was not called")
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
				t.Errorf("Access-Control-Allow-Origin = %q, want *", got)
			}
			if got := rec.Header().Get("Vary"); got != "" {
				t.Errorf("Vary = %q, want empty for the wildcard", got)
			}
		})
	}
}

func TestCORSEchoesAllowedOrigin(t *testing.T) {
	t.Parallel()

	mw := corsFromOrigins("https://app.example, https://admin.example")

	rec, nextHit := corsRequest(t, mw, http.MethodGet, "https://admin.example")
	if !nextHit {
		t.Fatal("next handler was not called")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://admin.example" {
		t.Errorf("Access-Control-Allow-Origin = %q, want the caller origin", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin", got)
	}
}

func TestCORSMismatchFallsBackToFirstConfiguredOrigin(t *testing.T) {
	t.Parallel()

	mw := corsFromOrigins("https://first.example, https://second.example")

	rec, nextHit := corsRequest(t, mw, http.MethodGet, "https://other.example")
	if !nextHit {
		t.Fatal("next handler was not called")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://first.example" {
		t.Errorf("Access-Control-Allow-Origin = %q, want the first configured origin", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	t.Parallel()

	rec, nextHit := corsRequest(t, corsFromOrigins(""), http.MethodOptions, "https://app.example")
	if nextHit {
		t.Fatal("preflight reached the wrapped handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, PATCH, DELETE, OPTIONS" {
		t.Errorf("Access-Control-Allow-Methods = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, X-Request-ID, X-User-ID, Authorization, X-API-Key" {
		t.Errorf("Access-Control-Allow-Headers = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Max-Age"); got != "600" {
		t.Errorf("Access-Control-Max-Age = %q, want 600", got)
	}
}

func TestCORSFromEnvReadsVariable(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://env.example")

	rec, _ := corsRequest(t, CORSFromEnv(), http.MethodGet, "https://env.example")
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://env.example" {
		t.Errorf("Access-Control-Allow-Origin = %q, want the configured origin", got)
	}
}
