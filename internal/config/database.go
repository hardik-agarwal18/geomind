package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func loadDatabase(lookup lookupEnv) (DatabaseConfig, error) {
	var cfg DatabaseConfig

	cfg.URL = get(lookup, "DATABASE_URL", "")

	if strings.TrimSpace(cfg.URL) == "" {
		return DatabaseConfig{}, fmt.Errorf(
			"DATABASE_URL is required",
		)
	}

	maxConns, err := parseInt32Setting(
		lookup,
		"DB_MAX_CONNS",
		"10",
		false,
	)
	if err != nil {
		return DatabaseConfig{}, err
	}

	minConns, err := parseInt32Setting(
		lookup,
		"DB_MIN_CONNS",
		"0",
		true,
	)
	if err != nil {
		return DatabaseConfig{}, err
	}

	if minConns > maxConns {
		return DatabaseConfig{}, fmt.Errorf(
			"DB_MIN_CONNS cannot exceed DB_MAX_CONNS",
		)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns

	durations := []struct {
		name     string
		fallback string
		target   *time.Duration
	}{
		{
			"DB_MAX_CONN_LIFETIME",
			"1h",
			&cfg.MaxConnLifetime,
		},
		{
			"DB_MAX_CONN_IDLE_TIME",
			"15m",
			&cfg.MaxConnIdleTime,
		},
		{
			"DB_HEALTH_CHECK_PERIOD",
			"1m",
			&cfg.HealthCheckPeriod,
		},
		{
			"DB_CONNECT_TIMEOUT",
			"5s",
			&cfg.ConnectTimeout,
		},
	}

	for _, item := range durations {
		value := get(
			lookup,
			item.name,
			item.fallback,
		)

		duration, err := time.ParseDuration(value)
		if err != nil {
			return DatabaseConfig{}, fmt.Errorf(
				"invalid %s: %w",
				item.name,
				err,
			)
		}

		if duration <= 0 {
			return DatabaseConfig{}, fmt.Errorf(
				"%s must be positive",
				item.name,
			)
		}

		*item.target = duration
	}

	return cfg, nil
}

func parseInt32Setting(
	lookup lookupEnv,
	name string,
	fallback string,
	allowZero bool,
) (int32, error) {

	value := get(lookup, name, fallback)

	parsed, err := strconv.ParseInt(
		value,
		10,
		32,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid %s: %w",
			name,
			err,
		)
	}

	if parsed < 0 || (!allowZero && parsed == 0) {
		return 0, fmt.Errorf(
			"%s has an invalid value",
			name,
		)
	}

	return int32(parsed), nil
}
