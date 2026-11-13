package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"geomind/internal/httpresponse"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d",
			http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("unexpected Content-Type: %q", got)
	}

	var response StatusResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected status ok, got %q", response.Status)
	}

	if response.Service != "geomind-api" {
		t.Errorf("unexpected service: %q", response.Service)
	}
}

func TestReadyHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	ReadyHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d",
			http.StatusOK, rec.Code)
	}

	var response StatusResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Status != "ready" {
		t.Errorf("expected ready, got %q", response.Status)
	}
}

func TestHandlersRejectUnsupportedMethods(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{"health", Handler, "/healthz"},
		{"readiness", ReadyHandler, "/readyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			rec := httptest.NewRecorder()

			tt.handler(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("expected 405, got %d", rec.Code)
			}

			if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
				t.Errorf("expected Allow: GET, got %q", got)
			}

			var response httpresponse.ErrorResponse

			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatalf("decode error response: %v", err)
			}

			if response.Error.Code != "method_not_allowed" {
				t.Errorf("unexpected error code: %q", response.Error.Code)
			}

			if response.Error.Message != "Only GET and HEAD are supported" {
				t.Errorf("unexpected error message: %q", response.Error.Message)
			}
		})
	}
}
