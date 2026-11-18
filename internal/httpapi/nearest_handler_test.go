package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"geomind/internal/places"
)

type nearestRepositoryStub struct {
	called  bool
	params  places.NearestParams
	results []places.NearestPlace
	err     error
}

func (s *nearestRepositoryStub) Nearest(
	ctx context.Context,
	params places.NearestParams,
) ([]places.NearestPlace, error) {
	s.called = true
	s.params = params

	return s.results, s.err
}

func newNearestTestHandler(
	repo *nearestRepositoryStub,
) *NearestPlacesHandler {
	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	service := places.NewNearestService(repo)

	return NewNearestPlacesHandler(service, logger)
}

func TestNearestHandlerDefaults(t *testing.T) {
	repo := &nearestRepositoryStub{}
	handler := newNearestTestHandler(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/places/nearest?latitude=12.9716&longitude=77.5946",
		nil,
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			rec.Code,
			rec.Body.String(),
		)
	}

	if !repo.called {
		t.Fatal("expected repository to be called")
	}

	if repo.params.K != places.DefaultNearestK {
		t.Errorf(
			"expected default K %d, got %d",
			places.DefaultNearestK,
			repo.params.K,
		)
	}

	if repo.params.CategorySlug != "" {
		t.Errorf(
			"expected no category filter, got %q",
			repo.params.CategorySlug,
		)
	}
}

func TestNearestHandlerParsesKAndCategory(t *testing.T) {
	repo := &nearestRepositoryStub{}
	handler := newNearestTestHandler(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/places/nearest?latitude=12.9716&longitude=77.5946&k=5&category=cafe",
		nil,
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			rec.Code,
			rec.Body.String(),
		)
	}

	if repo.params.K != 5 {
		t.Errorf("expected K=5, got %d", repo.params.K)
	}

	if repo.params.CategorySlug != "cafe" {
		t.Errorf(
			"expected category cafe, got %q",
			repo.params.CategorySlug,
		)
	}
}

func TestNearestHandlerRejectsInvalidParameters(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "missing latitude",
			query: "longitude=77.5946",
		},
		{
			name:  "missing longitude",
			query: "latitude=12.9716",
		},
		{
			name:  "invalid latitude",
			query: "latitude=91&longitude=77",
		},
		{
			name:  "invalid longitude",
			query: "latitude=12&longitude=181",
		},
		{
			name:  "zero K",
			query: "latitude=12&longitude=77&k=0",
		},
		{
			name:  "negative K",
			query: "latitude=12&longitude=77&k=-1",
		},
		{
			name:  "K above maximum",
			query: "latitude=12&longitude=77&k=101",
		},
		{
			name:  "non-integer K",
			query: "latitude=12&longitude=77&k=abc",
		},
		{
			name:  "invalid category",
			query: "latitude=12&longitude=77&category=Coffee!",
		},
		{
			name:  "unknown parameter",
			query: "latitude=12&longitude=77&debug=true",
		},
		{
			name:  "duplicate K",
			query: "latitude=12&longitude=77&k=1&k=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &nearestRepositoryStub{}
			handler := newNearestTestHandler(repo)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/places/nearest?"+tt.query,
				nil,
			)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected 400, got %d: %s",
					rec.Code,
					rec.Body.String(),
				)
			}

			if repo.called {
				t.Fatal(
					"repository must not be called for invalid input",
				)
			}
		})
	}
}

func TestNearestHandlerDeadline(t *testing.T) {
	repo := &nearestRepositoryStub{
		err: context.DeadlineExceeded,
	}

	handler := newNearestTestHandler(repo)
	handler.queryTimeout = 20 * time.Millisecond

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/places/nearest?latitude=12&longitude=77",
		nil,
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf(
			"expected 504, got %d",
			rec.Code,
		)
	}
}
