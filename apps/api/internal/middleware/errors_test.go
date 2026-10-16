package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteErrorUsesOneEnvelope(t *testing.T) {
	t.Parallel()

	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, r, http.StatusForbidden, CodeForbidden, "missing scope: write:contracts")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/contracts", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var env ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not the error envelope: %v (%s)", err, rec.Body.String())
	}
	if env.Error.Code != CodeForbidden {
		t.Errorf("error.code = %q, want %q", env.Error.Code, CodeForbidden)
	}
	if env.Error.Message != "missing scope: write:contracts" {
		t.Errorf("error.message = %q", env.Error.Message)
	}
	if env.Error.RequestID == "" {
		t.Error("error.request_id is empty; the envelope must carry the request ID")
	}
}

func TestErrorJSONCarriesRequestID(t *testing.T) {
	t.Parallel()

	var (
		body string
		rid  string
	)
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		body = ErrorJSON(r, CodeTimeout, "request exceeded the server time limit")
		rid = GetRequestID(r.Context())
	})
	RequestID(next).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	var env ErrorEnvelope
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("ErrorJSON produced invalid JSON: %v", err)
	}
	if env.Error.Code != CodeTimeout {
		t.Errorf("error.code = %q, want %q", env.Error.Code, CodeTimeout)
	}
	if rid == "" || env.Error.RequestID != rid {
		t.Errorf("error.request_id = %q, want the request context ID %q", env.Error.RequestID, rid)
	}
}
