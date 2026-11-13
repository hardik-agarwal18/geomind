package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"geomind/internal/health"
	"geomind/internal/httpresponse"
)

func newServerTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRoutes(t *testing.T) {
	handler := New(newServerTestLogger())

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
				if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
					t.Errorf("expected Allow: GET, HEAD, got %q", got)
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
	ts := httptest.NewServer(New(newServerTestLogger()))
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

func TestHeadHealthHasNoBody(t *testing.T) {
	ts := httptest.NewServer(New(newServerTestLogger()))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodHead, ts.URL+"/healthz", nil)
	if err != nil {
		t.Fatalf("create HEAD request: %v", err)
	}

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("send HEAD request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	if len(body) != 0 {
		t.Fatalf("HEAD response should have no body, got %q", body)
	}

	if resp.Header.Get("X-Request-ID") == "" {
		t.Error("expected request ID response header")
	}
}

func TestServeMuxSpecificity(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /places/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "place:%s", r.PathValue("id"))
	})

	mux.HandleFunc("GET /places/search", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "static-search-route")
	})

	tests := []struct {
		path string
		want string
	}{
		{"/places/search", "static-search-route"},
		{"/places/123", "place:123"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)

			mux.ServeHTTP(rec, req)

			if rec.Body.String() != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, rec.Body.String())
			}
		})
	}
}
