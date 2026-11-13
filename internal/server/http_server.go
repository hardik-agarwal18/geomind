package server

import (
	"net/http"

	"geomind/internal/config"
)

func NewHTTPServer(cfg config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: handler,

		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,

		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}
}
