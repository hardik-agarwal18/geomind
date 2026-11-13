package server

import (
	"log"
	"net/http"

	"geomind/internal/health"
	"geomind/internal/httpresponse"
)

func New() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", health.Handler)
	mux.HandleFunc("/readyz", health.ReadyHandler)

	// Catch unmatched paths and return a consistent JSON 404.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := httpresponse.WriteError(
			w,
			http.StatusNotFound,
			"not_found",
			"Resource not found",
		); err != nil {
			log.Printf("failed to write not found response: %v", err)
		}
	})

	return mux
}
