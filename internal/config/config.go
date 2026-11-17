package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv          string
	HTTPAddr        string
	LogLevel        slog.Level
	ShutdownTimeout time.Duration

	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int

	Database DatabaseConfig
}

type lookupEnv func(string) (string, bool)

// Load reads environment variables, applies defaults, and validates the result.
func Load() (Config, error) {
	return load(os.LookupEnv)
}

func load(lookup lookupEnv) (Config, error) {
	cfg := Config{
		AppEnv:   get(lookup, "APP_ENV", "development"),
		HTTPAddr: get(lookup, "HTTP_ADDR", ":8080"),
	}

	switch cfg.AppEnv {
	case "development", "test", "production":
	default:
		return Config{}, fmt.Errorf(
			"invalid APP_ENV %q: expected development, test, or production",
			cfg.AppEnv,
		)
	}

	level, err := parseLogLevel(get(lookup, "LOG_LEVEL", "info"))
	if err != nil {
		return Config{}, err
	}
	cfg.LogLevel = level

	durations := []struct {
		name     string
		fallback string
		target   *time.Duration
	}{
		{"SHUTDOWN_TIMEOUT", "10s", &cfg.ShutdownTimeout},
		{"READ_HEADER_TIMEOUT", "5s", &cfg.ReadHeaderTimeout},
		{"READ_TIMEOUT", "15s", &cfg.ReadTimeout},
		{"WRITE_TIMEOUT", "30s", &cfg.WriteTimeout},
		{"IDLE_TIMEOUT", "60s", &cfg.IdleTimeout},
	}

	for _, item := range durations {
		value := get(lookup, item.name, item.fallback)

		duration, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"invalid %s %q: %w",
				item.name,
				value,
				err,
			)
		}

		if duration <= 0 {
			return Config{}, fmt.Errorf(
				"%s must be greater than zero",
				item.name,
			)
		}

		*item.target = duration
	}

	maxHeaderBytes := get(lookup, "MAX_HEADER_BYTES", "1048576")

	cfg.MaxHeaderBytes, err = strconv.Atoi(maxHeaderBytes)
	if err != nil {
		return Config{}, fmt.Errorf(
			"invalid MAX_HEADER_BYTES %q: %w",
			maxHeaderBytes,
			err,
		)
	}

	if cfg.MaxHeaderBytes <= 0 {
		return Config{}, fmt.Errorf(
			"MAX_HEADER_BYTES must be greater than zero",
		)
	}

	if err := validateHTTPAddr(cfg.HTTPAddr); err != nil {
		return Config{}, fmt.Errorf("invalid HTTP_ADDR: %w", err)
	}

	// Load and validate database configuration.
	databaseConfig, err := loadDatabase(lookup)
	if err != nil {
		return Config{}, fmt.Errorf(
			"load database configuration: %w",
			err,
		)
	}

	cfg.Database = databaseConfig

	return cfg, nil
}

func get(lookup lookupEnv, name, fallback string) string {
	value, ok := lookup(name)

	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}

	return strings.TrimSpace(value)
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf(
			"invalid LOG_LEVEL %q: expected debug, info, warn, or error",
			value,
		)
	}
}

func validateHTTPAddr(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("expected host:port format: %w", err)
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("port must be numeric: %w", err)
	}

	if portNumber < 0 || portNumber > 65535 {
		return fmt.Errorf("port must be between 0 and 65535")
	}

	return nil
}

type DatabaseConfig struct {
	URL string

	MaxConns int32
	MinConns int32

	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration

	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}
