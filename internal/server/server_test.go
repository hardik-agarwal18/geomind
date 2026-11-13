package server

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"geomind/internal/health"
	"geomind/internal/httpresponse"
)

func TestRoutes(t *testing.T) {
	handler := New()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"health", http.MethodGet, "/healthz", http.StatusOK},
		{"readiness", http.MethodGet, "/readyz", http.StatusOK},
		{"unsupported method", http.MethodPost, "/healthz", http.StatusMethodNotAllowed},
		{"unknown route", http.MethodGet, "/unknown", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d",
					tt.wantStatus, rec.Code)
			}

			if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Errorf("unexpected Content-Type: %q", got)
			}

			if tt.wantStatus == http.StatusMethodNotAllowed {
				if got := rec.Header().Get("Allow"); got != http.MethodGet {
					t.Errorf("expected Allow: GET, got %q", got)
				}
			}

			switch tt.wantStatus {
			case http.StatusOK:
				var response health.StatusResponse
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("decode success response: %v", err)
				}

			case http.StatusMethodNotAllowed:
				var response httpresponse.ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("decode method error: %v", err)
				}
				if response.Error.Code != "method_not_allowed" {
					t.Errorf("unexpected error code: %q", response.Error.Code)
				}

			case http.StatusNotFound:
				var response httpresponse.ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("decode not found error: %v", err)
				}
				if response.Error.Code != "not_found" {
					t.Errorf("unexpected error code: %q", response.Error.Code)
				}
			}
		})
	}
}

// Malformed HTTP syntax is rejected by net/http before a handler is called.
func TestMalformedHTTPRequest(t *testing.T) {
	ts := httptest.NewServer(New())
	defer ts.Close()

	address := strings.TrimPrefix(ts.URL, "http://")

	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		t.Fatalf("connect to test server: %v", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("set connection deadline: %v", err)
	}

	_, err = io.WriteString(
		conn,
		"GET /healthz INVALID\r\nHost: localhost\r\n\r\n",
	)
	if err != nil {
		t.Fatalf("write malformed request: %v", err)
	}

	reader := bufio.NewReader(conn)

	statusLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read response status line: %v", err)
	}

	if !strings.Contains(statusLine, "400 Bad Request") {
		t.Fatalf("expected 400 Bad Request, got %q", statusLine)
	}
}
