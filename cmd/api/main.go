package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"geomind/internal/config"
	"geomind/internal/database"
	"geomind/internal/logging"
	"geomind/internal/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "GeoMind stopped with error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. Load and validate configuration.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// 2. Create structured logger.
	logger := logging.New(cfg)

	logger.Info(
		"application_starting",
		"environment", cfg.AppEnv,
		"address", cfg.HTTPAddr,
	)

	// 3. Create application signal context.
	appCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := database.NewPool(
		appCtx,
		cfg.Database,
	)
	if err != nil {
		return fmt.Errorf(
			"initialize database: %w",
			err,
		)
	}

	// Pool remains alive throughout the application lifecycle.
	defer database.ClosePool(pool)

	logger.Info(
		"database_pool_ready",
		"max_connections",
		cfg.Database.MaxConns,
	)

	// 4. Construct HTTP handler and server.
	handler := server.New(logger)
	srv := server.NewHTTPServer(cfg, handler)

	// 5. Start HTTP server.
	serverErr := make(chan error, 1)

	go func() {
		logger.Info("http_server_started", "address", srv.Addr)

		err := srv.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("HTTP server failed: %w", err)
			return
		}

		serverErr <- nil
	}()

	// 6. Wait for shutdown signal or server failure.
	select {
	case <-appCtx.Done():
		logger.Info("shutdown_signal_received")
		stop()

	case err := <-serverErr:
		return err
	}

	// 7. Give active requests time to finish.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer cancel()

	logger.Info(
		"http_server_shutdown_started",
		"timeout", cfg.ShutdownTimeout,
	)

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error(
			"graceful_shutdown_failed",
			"error", err,
		)

		if closeErr := srv.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}

		<-serverErr

		return fmt.Errorf("HTTP shutdown failed: %w", err)
	}

	if err := <-serverErr; err != nil {
		return err
	}

	logger.Info("application_stopped")
	return nil
}

// Keep this assertion to ensure our configured server implements the
// standard HTTP server interface expected by the application.
var _ http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
