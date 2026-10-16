package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sorolens/sorolens/apps/api/internal/store"
)

// metricsRequest runs mw against GET /metrics and reports whether the wrapped
// handler was reached.
func metricsRequest(t *testing.T, mw func(http.Handler) http.Handler, prepare func(*http.Request)) (bool, *httptest.ResponseRecorder) {
	t.Helper()

	hit := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hit = true })
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	if prepare != nil {
		prepare(req)
	}
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)
	return hit, rec
}

func TestProtectMetricsOpenWithoutToken(t *testing.T) {
	t.Parallel()

	hit, rec := metricsRequest(t, protectMetrics("", store.NewMockStore(), discardLogger()), nil)
	if !hit {
		t.Fatal("request without a configured token did not reach the handler")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestProtectMetricsRejectsMissingCredential(t *testing.T) {
	t.Parallel()

	hit, rec := metricsRequest(t, protectMetrics("s3cret", store.NewMockStore(), discardLogger()), nil)
	if hit {
		t.Fatal("request without credentials reached the handler")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
		t.Errorf("WWW-Authenticate = %q, want Bearer", got)
	}
}

func TestProtectMetricsRejectsWrongToken(t *testing.T) {
	t.Parallel()

	hit, rec := metricsRequest(t, protectMetrics("s3cret", store.NewMockStore(), discardLogger()), func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer nope")
	})
	if hit {
		t.Fatal("request with the wrong token reached the handler")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestProtectMetricsAcceptsBearerToken(t *testing.T) {
	t.Parallel()

	for _, header := range []string{"Bearer s3cret", "bearer s3cret"} {
		hit, _ := metricsRequest(t, protectMetrics("s3cret", store.NewMockStore(), discardLogger()), func(r *http.Request) {
			r.Header.Set("Authorization", header)
		})
		if !hit {
			t.Errorf("Authorization %q did not reach the handler", header)
		}
	}
}

func TestProtectMetricsAcceptsAdminIdentity(t *testing.T) {
	t.Parallel()

	ms := store.NewMockStore()
	ms.AddUser(store.User{ID: "admin-1", Role: store.RoleAdmin})

	hit, _ := metricsRequest(t, protectMetrics("s3cret", ms, discardLogger()), func(r *http.Request) {
		r.Header.Set("X-User-ID", "admin-1")
	})
	if !hit {
		t.Error("admin identity did not reach the handler")
	}
}

func TestProtectMetricsRejectsNonAdminIdentity(t *testing.T) {
	t.Parallel()

	ms := store.NewMockStore()
	ms.AddUser(store.User{ID: "viewer-1", Role: store.RoleViewer})

	hit, rec := metricsRequest(t, protectMetrics("s3cret", ms, discardLogger()), func(r *http.Request) {
		r.Header.Set("X-User-ID", "viewer-1")
	})
	if hit {
		t.Fatal("viewer identity reached the handler")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestProtectMetricsFromEnvReadsVariable(t *testing.T) {
	t.Setenv("METRICS_TOKEN", "from-env")

	hit, rec := metricsRequest(t, ProtectMetricsFromEnv(store.NewMockStore(), discardLogger()), nil)
	if hit {
		t.Fatal("request without credentials reached the handler")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
