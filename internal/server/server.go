package server

import (
	"log/slog"
	"net/http"

	"geomind/internal/health"
	"geomind/internal/httpresponse"
	"geomind/internal/middleware"
)

func New(logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", health.Handler)
	mux.HandleFunc("GET /readyz", health.ReadyHandler)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := httpresponse.WriteMethodNotAllowed(w); err != nil {
			logger.Error("failed to write health method error", "error", err)
		}
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := httpresponse.WriteMethodNotAllowed(w); err != nil {
			logger.Error("failed to write readiness method error", "error", err)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := httpresponse.WriteError(
			w,
			http.StatusNotFound,
			"not_found",
			"Resource not found",
		); err != nil {
			logger.Error("failed to write not found response", "error", err)
		}
	})

	return middleware.Chain(
		mux,
		middleware.RequestID(logger),
		middleware.AccessLog(logger),
		middleware.Recovery(logger),
	)
}
