package server

import (
	"log"
	"net/http"

	"geomind/internal/health"
	"geomind/internal/httpresponse"
	"geomind/internal/middleware"
)

func New(logger *log.Logger) http.Handler {
	if logger == nil {
		logger = log.Default()
	}

	mux := http.NewServeMux()

	// Method-aware routes.
	mux.HandleFunc("GET /healthz", health.Handler)
	mux.HandleFunc("GET /readyz", health.ReadyHandler)

	// Exact-path fallbacks for unsupported methods.
	// GET-specific patterns take precedence for GET and HEAD.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := httpresponse.WriteMethodNotAllowed(w); err != nil {
			logger.Printf("failed to write health method error: %v", err)
		}
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := httpresponse.WriteMethodNotAllowed(w); err != nil {
			logger.Printf("failed to write readiness method error: %v", err)
		}
	})

	// Catch-all for unknown paths.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := httpresponse.WriteError(
			w,
			http.StatusNotFound,
			"not_found",
			"Resource not found",
		); err != nil {
			logger.Printf("failed to write not found response: %v", err)
		}
	})

	// Request execution order: RequestID -> AccessLog -> Recovery -> ServeMux.
	return middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.AccessLog(logger),
		middleware.Recovery(logger),
	)
}
