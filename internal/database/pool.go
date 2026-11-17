package database

import (
	"context"
	"fmt"
	"time"

	"geomind/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates, configures and verifies a PostgreSQL pool.
func NewPool(
	ctx context.Context,
	cfg config.DatabaseConfig,
) (*pgxpool.Pool, error) {

	if cfg.URL == "" {
		return nil, fmt.Errorf(
			"database URL is required",
		)
	}

	if cfg.MaxConns <= 0 {
		return nil, fmt.Errorf(
			"maximum connections must be positive",
		)
	}

	if cfg.MinConns < 0 ||
		cfg.MinConns > cfg.MaxConns {
		return nil, fmt.Errorf(
			"invalid minimum connection count",
		)
	}

	parsedConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf(
			"parse database connection configuration: %w",
			err,
		)
	}

	// Pool configuration.
	parsedConfig.MaxConns = cfg.MaxConns
	parsedConfig.MinConns = cfg.MinConns

	parsedConfig.MaxConnLifetime =
		cfg.MaxConnLifetime

	parsedConfig.MaxConnIdleTime =
		cfg.MaxConnIdleTime

	parsedConfig.HealthCheckPeriod =
		cfg.HealthCheckPeriod

	// Connection establishment timeout.
	parsedConfig.ConnConfig.ConnectTimeout =
		cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(
		ctx,
		parsedConfig,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create PostgreSQL pool: %w",
			err,
		)
	}

	// NewWithConfig does not guarantee that a
	// connection can actually be established.
	pingCtx, cancel := context.WithTimeout(
		ctx,
		cfg.ConnectTimeout,
	)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()

		return nil, fmt.Errorf(
			"PostgreSQL startup ping failed: %w",
			err,
		)
	}

	return pool, nil
}

// ClosePool provides a simple explicit lifecycle helper.
func ClosePool(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}

// Keep time imported here for the package's configuration
// documentation and future database lifecycle extensions.
var _ = time.Second
