package database

import (
	"context"
	"testing"
	"time"

	"geomind/internal/config"
)

func TestNewPoolFailsWhenDatabaseUnavailable(t *testing.T) {
	cfg := config.DatabaseConfig{
		URL: "postgres://invalid:invalid@127.0.0.1:1/geomind?sslmode=disable",

		MaxConns: 1,
		MinConns: 0,

		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   time.Minute,
		HealthCheckPeriod: time.Minute,
		ConnectTimeout:    500 * time.Millisecond,
	}

	pool, err := NewPool(
		context.Background(),
		cfg,
	)

	if err == nil {
		if pool != nil {
			pool.Close()
		}

		t.Fatal("expected connection failure")
	}

	if pool != nil {
		t.Fatal("expected nil pool on initialization failure")
	}
}
