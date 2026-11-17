package server

import (
	"log/slog"
	"net/http"

	"geomind/internal/health"
	"geomind/internal/httpresponse"
	"geomind/internal/middleware"
)

// New creates the existing server without optional API routes.
func New(logger *slog.Logger) http.Handler {
	return newServer(logger, nil)
}

// NewWithNearbyPlaces creates the server with the nearby places route.
func NewWithNearbyPlaces(
	logger *slog.Logger,
	nearbyHandler http.Handler,
) http.Handler {
	return newServer(logger, nearbyHandler)
}

func newServer(
	logger *slog.Logger,
	nearbyHandler http.Handler,
) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()

	// Health and readiness.
	mux.HandleFunc("GET /healthz", health.Handler)
	mux.HandleFunc("GET /readyz", health.ReadyHandler)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := httpresponse.WriteMethodNotAllowed(w); err != nil {
			logger.Error(
				"failed to write health method error",
				"error", err,
			)
		}
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := httpresponse.WriteMethodNotAllowed(w); err != nil {
			logger.Error(
				"failed to write readiness method error",
				"error", err,
			)
		}
	})

	// Nearby places API.
	if nearbyHandler != nil {
		mux.Handle(
			"GET /api/v1/places/nearby",
			nearbyHandler,
		)

		mux.HandleFunc(
			"/api/v1/places/nearby",
			func(w http.ResponseWriter, r *http.Request) {
				if err := httpresponse.WriteMethodNotAllowed(w); err != nil {
					logger.Error(
						"failed to write nearby method error",
						"error", err,
					)
				}
			},
		)
	}

	// Not found.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := httpresponse.WriteError(
			w,
			http.StatusNotFound,
			"not_found",
			"Resource not found",
		); err != nil {
			logger.Error(
				"failed to write not found response",
				"error", err,
			)
		}
	})

	return middleware.Chain(
		mux,
		middleware.RequestID(logger),
		middleware.AccessLog(logger),
		middleware.Recovery(logger),
	)
}
