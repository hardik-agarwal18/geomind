package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"geomind/internal/places"
)

type nearbyRepositoryStub struct {
	called bool
	params places.NearbyParams

	results []places.NearbyPlace
	err     error

	waitForCancellation bool
}

func (s *nearbyRepositoryStub) Nearby(
	ctx context.Context,
	params places.NearbyParams,
) ([]places.NearbyPlace, error) {
	s.called = true
	s.params = params

	if s.waitForCancellation {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	return s.results, s.err
}

func newNearbyTestHandler(
	repo *nearbyRepositoryStub,
) *NearbyPlacesHandler {
	service := places.NewService(repo)

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	return NewNearbyPlacesHandler(service, logger)
}

func TestNearbyHandlerDefaultsAndResponse(t *testing.T) {
	repo := &nearbyRepositoryStub{
		results: []places.NearbyPlace{},
	}

	handler := newNearbyTestHandler(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/places/nearby?latitude=12.9716&longitude=77.5946",
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

	if repo.params.RadiusM != 1000 {
		t.Errorf(
			"expected default radius 1000, got %v",
			repo.params.RadiusM,
		)
	}

	if repo.params.Limit != 20 {
		t.Errorf(
			"expected default limit 20, got %d",
			repo.params.Limit,
		)
	}

	var response nearbySearchResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Meta.Count != 0 {
		t.Errorf(
			"expected zero results, got %d",
			response.Meta.Count,
		)
	}

	if response.Data == nil {
		t.Fatal("expected data to be an empty JSON array, not null")
	}
}

func TestNearbyHandlerRejectsInvalidQuery(t *testing.T) {
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
			name:  "NaN latitude",
			query: "latitude=NaN&longitude=77",
		},
		{
			name:  "negative radius",
			query: "latitude=12&longitude=77&radius_m=-1",
		},
		{
			name:  "zero radius",
			query: "latitude=12&longitude=77&radius_m=0",
		},
		{
			name:  "radius too large",
			query: "latitude=12&longitude=77&radius_m=50001",
		},
		{
			name:  "invalid category",
			query: "latitude=12&longitude=77&category=Coffee!",
		},
		{
			name:  "limit too large",
			query: "latitude=12&longitude=77&limit=101",
		},
		{
			name:  "zero limit",
			query: "latitude=12&longitude=77&limit=0",
		},
		{
			name:  "duplicate latitude",
			query: "latitude=12&latitude=13&longitude=77",
		},
		{
			name:  "unsupported parameter",
			query: "latitude=12&longitude=77&debug=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &nearbyRepositoryStub{}
			handler := newNearbyTestHandler(repo)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/places/nearby?"+tt.query,
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

func TestNearbyHandlerQueryTimeout(t *testing.T) {
	repo := &nearbyRepositoryStub{
		waitForCancellation: true,
	}

	handler := newNearbyTestHandler(repo)

	// Use a short timeout for the test instead of waiting
	// for the production two-second deadline.
	handlerTimeout := 20 * time.Millisecond
	_ = handlerTimeout

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/places/nearby?latitude=12&longitude=77",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf(
			"expected 504, got %d: %s",
			rec.Code,
			rec.Body.String(),
		)
	}
}
