package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"geomind/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger := log.Default()

	// Context representing the application's lifetime.
	appCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	handler := server.New(logger)
	srv := server.NewHTTPServer(":8080", handler)

	serverErr := make(chan error, 1)

	// Start the HTTP server in a separate goroutine.
	go func() {
		logger.Printf("GeoMind API listening on %s", srv.Addr)

		err := srv.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("HTTP server failed: %w", err)
			return
		}

		serverErr <- nil
	}()

	select {
	case <-appCtx.Done():
		logger.Println("Shutdown signal received")

		// Restore default signal handling.
		// A second Ctrl+C can force termination.
		stop()

	case err := <-serverErr:
		return err
	}

	// Give active HTTP requests up to 10 seconds to finish.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	logger.Println("Starting graceful HTTP shutdown")

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("Graceful shutdown failed: %v", err)

		// Force-close connections if the deadline expires.
		if closeErr := srv.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}

		// Wait for the server goroutine to exit.
		<-serverErr

		return fmt.Errorf("HTTP shutdown failed: %w", err)
	}

	if err := <-serverErr; err != nil {
		return err
	}

	logger.Println("GeoMind API stopped successfully")
	return nil
}
