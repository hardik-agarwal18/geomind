package config

import (
	"log/slog"
	"testing"
	"time"
)

func lookupFrom(values map[string]string) lookupEnv {
	defaults := map[string]string{
		"DATABASE_URL": "postgres://test:test@localhost:5433/geomind",
	}

	// Test-specific settings override defaults.
	for key, value := range values {
		defaults[key] = value
	}

	return func(key string) (string, bool) {
		value, ok := defaults[key]
		return value, ok
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := load(lookupFrom(map[string]string{}))
	if err != nil {
		t.Fatalf("load defaults: %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Errorf("expected development, got %q", cfg.AppEnv)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Errorf("expected :8080, got %q", cfg.HTTPAddr)
	}

	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("expected info level, got %v", cfg.LogLevel)
	}

	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("unexpected shutdown timeout: %s", cfg.ShutdownTimeout)
	}

	if cfg.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("unexpected read header timeout: %s", cfg.ReadHeaderTimeout)
	}

	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("unexpected read timeout: %s", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout != 30*time.Second {
		t.Errorf("unexpected write timeout: %s", cfg.WriteTimeout)
	}

	if cfg.IdleTimeout != 60*time.Second {
		t.Errorf("unexpected idle timeout: %s", cfg.IdleTimeout)
	}

	if cfg.MaxHeaderBytes != 1<<20 {
		t.Errorf("unexpected maximum header size: %d", cfg.MaxHeaderBytes)
	}
}

func TestLoadCustomValues(t *testing.T) {
	values := map[string]string{
		"APP_ENV":             "production",
		"HTTP_ADDR":           "127.0.0.1:9090",
		"LOG_LEVEL":           "debug",
		"SHUTDOWN_TIMEOUT":    "20s",
		"READ_HEADER_TIMEOUT": "3s",
		"READ_TIMEOUT":        "12s",
		"WRITE_TIMEOUT":       "25s",
		"IDLE_TIMEOUT":        "45s",
		"MAX_HEADER_BYTES":    "2097152",
	}

	cfg, err := load(lookupFrom(values))
	if err != nil {
		t.Fatalf("load custom config: %v", err)
	}

	if cfg.AppEnv != "production" {
		t.Errorf("unexpected environment: %q", cfg.AppEnv)
	}

	if cfg.HTTPAddr != "127.0.0.1:9090" {
		t.Errorf("unexpected address: %q", cfg.HTTPAddr)
	}

	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("unexpected log level: %v", cfg.LogLevel)
	}

	if cfg.ShutdownTimeout != 20*time.Second {
		t.Errorf("unexpected shutdown timeout: %s", cfg.ShutdownTimeout)
	}

	if cfg.MaxHeaderBytes != 2<<20 {
		t.Errorf("unexpected max header bytes: %d", cfg.MaxHeaderBytes)
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]string
	}{
		{
			name:   "invalid environment",
			values: map[string]string{"APP_ENV": "staging"},
		},
		{
			name:   "invalid log level",
			values: map[string]string{"LOG_LEVEL": "banana"},
		},
		{
			name:   "invalid address",
			values: map[string]string{"HTTP_ADDR": "not-an-address"},
		},
		{
			name:   "invalid port",
			values: map[string]string{"HTTP_ADDR": ":abc"},
		},
		{
			name:   "invalid duration",
			values: map[string]string{"SHUTDOWN_TIMEOUT": "tomorrow"},
		},
		{
			name:   "zero timeout",
			values: map[string]string{"READ_TIMEOUT": "0s"},
		},
		{
			name:   "negative timeout",
			values: map[string]string{"WRITE_TIMEOUT": "-1s"},
		},
		{
			name:   "invalid header size",
			values: map[string]string{"MAX_HEADER_BYTES": "large"},
		},
		{
			name:   "zero header size",
			values: map[string]string{"MAX_HEADER_BYTES": "0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := load(lookupFrom(tt.values))
			if err == nil {
				t.Fatal("expected configuration validation error")
			}
		})
	}
}
