package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"geomind/internal/httpresponse"
)

func TestChainExecutionOrder(t *testing.T) {
	var events []string

	record := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				events = append(events, name+":before")

				next.ServeHTTP(w, r)

				events = append(events, name+":after")
			})
		}
	}

	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		events = append(events, "handler")
	})

	handler := Chain(endpoint, record("first"), record("second"))

	handler.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	want := []string{
		"first:before",
		"second:before",
		"handler",
		"second:after",
		"first:after",
	}

	if len(events) != len(want) {
		t.Fatalf("expected events %v, got %v", want, events)
	}

	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("expected events %v, got %v", want, events)
		}
	}
}

func TestRequestIDPropagation(t *testing.T) {
	var contextID string

	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID = RequestIDFromContext(r.Context())

		w.WriteHeader(http.StatusNoContent)
	})

	handler := RequestID(endpoint)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	headerID := rec.Header().Get(RequestIDHeader)

	if headerID == "" {
		t.Fatal("expected request ID response header")
	}

	if contextID == "" {
		t.Fatal("expected request ID in request context")
	}

	if headerID != contextID {
		t.Fatalf("header ID %q differs from context ID %q", headerID, contextID)
	}
}

func TestAccessLogCapturesStatus(t *testing.T) {
	tests := []struct {
		name       string
		explicit   bool
		wantStatus string
	}{
		{
			name:       "implicit 200",
			explicit:   false,
			wantStatus: "status=200",
		},
		{
			name:       "explicit 202",
			explicit:   true,
			wantStatus: "status=202",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := log.New(&logs, "", 0)

			endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.explicit {
					w.WriteHeader(http.StatusAccepted)
					return
				}

				_, _ = io.WriteString(w, "private-response-body")
			})

			handler := Chain(endpoint, RequestID, AccessLog(logger))

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/example", nil)

			handler.ServeHTTP(rec, req)

			if rec.Code == 0 {
				t.Fatal("expected a committed HTTP response")
			}

			logOutput := logs.String()

			if !strings.Contains(logOutput, tt.wantStatus) {
				t.Fatalf("expected log to contain %q, got %q",
					tt.wantStatus, logOutput)
			}

			requestID := rec.Header().Get(RequestIDHeader)

			if requestID == "" ||
				!strings.Contains(logOutput, "request_id="+requestID) {
				t.Fatalf("expected request ID in access log, got %q", logOutput)
			}

			if strings.Contains(logOutput, "private-response-body") {
				t.Fatal("access log must not contain response body")
			}

			if !strings.Contains(logOutput, `method="GET"`) {
				t.Errorf("expected HTTP method in access log, got %q", logOutput)
			}

			if !strings.Contains(logOutput, `path="/example"`) {
				t.Errorf("expected request path in access log, got %q", logOutput)
			}
		})
	}
}

func TestRecoveryReturnsGenericInternalError(t *testing.T) {
	var logs bytes.Buffer
	logger := log.New(&logs, "", 0)

	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("private-panic-value")
	})

	handler := Chain(
		endpoint,
		RequestID,
		AccessLog(logger),
		Recovery(logger),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	requestID := rec.Header().Get(RequestIDHeader)

	if requestID == "" {
		t.Fatal("request ID header was lost during recovery")
	}

	var response httpresponse.ErrorResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode recovery response: %v", err)
	}

	if response.Error.Code != "internal_error" {
		t.Errorf("unexpected error code: %q", response.Error.Code)
	}

	if response.Error.Message != "Internal server error" {
		t.Errorf("unexpected error message: %q", response.Error.Message)
	}

	logOutput := logs.String()

	if !strings.Contains(logOutput, "panic_recovered") {
		t.Error("expected panic recovery log")
	}

	if !strings.Contains(logOutput, "status=500") {
		t.Error("expected access log to capture recovered 500")
	}

	if !strings.Contains(logOutput, "request_id="+requestID) {
		t.Error("expected request ID in panic logs")
	}

	if strings.Contains(rec.Body.String(), "private-panic-value") ||
		strings.Contains(logOutput, "private-panic-value") {
		t.Fatal("panic value should not be exposed")
	}
}

func TestResponseControllerSurvivesMiddleware(t *testing.T) {
	logger := log.New(io.Discard, "", 0)

	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Handler-Header", "preserved")

		controller := http.NewResponseController(w)

		if err := controller.Flush(); err != nil {
			t.Errorf("flush through middleware: %v", err)
		}
	})

	handler := Chain(
		endpoint,
		RequestID,
		AccessLog(logger),
		Recovery(logger),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Handler-Header"); got != "preserved" {
		t.Errorf("handler response header was lost: %q", got)
	}

	if got := rec.Header().Get(RequestIDHeader); got == "" {
		t.Error("request ID response header was lost")
	}

	if !rec.Flushed {
		t.Error("expected response to be flushed through middleware")
	}
}
