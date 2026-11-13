package server

import (
	"net/http"
	"testing"
	"time"

	"geomind/internal/config"
)

func TestNewHTTPServerTimeoutConfiguration(t *testing.T) {
	cfg := config.Config{
		HTTPAddr:          ":8080",
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	srv := NewHTTPServer(cfg, http.NotFoundHandler())

	if srv.Addr != ":8080" {
		t.Errorf("expected address :8080, got %q", srv.Addr)
	}

	tests := []struct {
		name string
		got  time.Duration
		want time.Duration
	}{
		{"read header timeout", srv.ReadHeaderTimeout, 5 * time.Second},
		{"read timeout", srv.ReadTimeout, 15 * time.Second},
		{"write timeout", srv.WriteTimeout, 30 * time.Second},
		{"idle timeout", srv.IdleTimeout, 60 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("expected %s, got %s", tt.want, tt.got)
			}
		})
	}

	if srv.MaxHeaderBytes != 1<<20 {
		t.Errorf("expected MaxHeaderBytes = 1 MiB, got %d", srv.MaxHeaderBytes)
	}
}
