package server

import (
	"net/http"

	"geomind/internal/health"
)

func New() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", health.Handler)

	return mux
}
