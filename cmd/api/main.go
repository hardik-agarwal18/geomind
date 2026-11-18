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
	"geomind/internal/httpapi"
	"geomind/internal/logging"
	"geomind/internal/places"
	"geomind/internal/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"GeoMind stopped with error: %v\n",
			err,
		)

		os.Exit(1)
	}
}

func run() error {
	// --------------------------------------------------
	// 1. Load and validate configuration
	// --------------------------------------------------

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf(
			"load configuration: %w",
			err,
		)
	}

	// --------------------------------------------------
	// 2. Initialize structured logger
	// --------------------------------------------------

	logger := logging.New(cfg)

	logger.Info(
		"application_starting",
		"service", "geomind-api",
		"environment", cfg.AppEnv,
		"address", cfg.HTTPAddr,
	)

	// --------------------------------------------------
	// 3. Create application signal context
	// --------------------------------------------------

	appCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// --------------------------------------------------
	// 4. Initialize PostgreSQL connection pool
	// --------------------------------------------------

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

	// The pool remains alive until HTTP shutdown completes.
	defer database.ClosePool(pool)

	logger.Info(
		"database_pool_ready",
		"max_connections", cfg.Database.MaxConns,
		"min_connections", cfg.Database.MinConns,
	)

	// --------------------------------------------------
	// 5. Initialize repositories
	// --------------------------------------------------

	placeRepository := places.NewRepository(pool)

	logger.Info("repositories_initialized")

	// --------------------------------------------------
	// 6. Initialize application services
	// --------------------------------------------------

	// Radius-based nearby search.
	nearbyService := places.NewService(
		placeRepository,
	)

	// Nearest-K search.
	nearestService := places.NewNearestService(
		placeRepository,
	)

	logger.Info("application_services_initialized")

	// --------------------------------------------------
	// 7. Initialize HTTP handlers
	// --------------------------------------------------

	nearbyHandler := httpapi.NewNearbyPlacesHandler(
		nearbyService,
		logger,
	)

	nearestHandler := httpapi.NewNearestPlacesHandler(
		nearestService,
		logger,
	)

	// --------------------------------------------------
	// 8. Construct HTTP routes and middleware
	// --------------------------------------------------

	handler := server.NewWithPlaceSearchHandlers(
		logger,
		nearbyHandler,
		nearestHandler,
	)

	// --------------------------------------------------
	// 9. Construct HTTP server
	// --------------------------------------------------

	srv := server.NewHTTPServer(
		cfg,
		handler,
	)

	// --------------------------------------------------
	// 10. Start HTTP server
	// --------------------------------------------------

	serverErr := make(chan error, 1)

	go func() {
		logger.Info(
			"http_server_started",
			"address", srv.Addr,
		)

		err := srv.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf(
				"HTTP server failed: %w",
				err,
			)
			return
		}

		serverErr <- nil
	}()

	// --------------------------------------------------
	// 11. Wait for shutdown signal or server failure
	// --------------------------------------------------

	select {
	case <-appCtx.Done():
		logger.Info("shutdown_signal_received")

		// Stop receiving additional termination signals
		// through this signal context.
		stop()

	case err := <-serverErr:
		return err
	}

	// --------------------------------------------------
	// 12. Graceful HTTP shutdown
	// --------------------------------------------------

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

		// Force-close remaining connections if graceful
		// shutdown cannot complete within the deadline.
		if closeErr := srv.Close(); closeErr != nil {
			<-serverErr

			return errors.Join(err, closeErr)
		}

		<-serverErr

		return fmt.Errorf(
			"HTTP shutdown failed: %w",
			err,
		)
	}

	// Wait for ListenAndServe to finish.
	if err := <-serverErr; err != nil {
		return err
	}

	logger.Info("http_server_stopped")

	// The deferred database.ClosePool executes after this
	// function returns, so active HTTP handlers get the
	// opportunity to finish their database operations first.

	logger.Info("application_stopped")

	return nil
}
