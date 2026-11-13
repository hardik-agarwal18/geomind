package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestContextCancelledOnClientDisconnect(t *testing.T) {
	started := make(chan struct{})
	cancelled := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)

		<-r.Context().Done()

		close(cancelled)
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		ts.URL,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	clientDone := make(chan struct{})

	go func() {
		defer close(clientDone)

		resp, err := ts.Client().Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("handler did not start")
	}

	// Simulate the client abandoning the request.
	cancel()

	select {
	case <-cancelled:
	case <-time.After(3 * time.Second):
		t.Fatal("handler did not observe cancellation")
	}

	select {
	case <-clientDone:
	case <-time.After(3 * time.Second):
		t.Fatal("client request did not finish")
	}
}

func TestGracefulShutdownCompletes(t *testing.T) {
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	serveDone := make(chan error, 1)

	go func() {
		serveDone <- srv.Serve(listener)
	}()

	// Allow the server to enter its serving loop.
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get("http://" + listener.Addr().String())
	if err != nil {
		t.Fatalf("server request failed: %v", err)
	}
	resp.Body.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	start := time.Now()

	err = srv.Shutdown(ctx)

	if err != nil {
		t.Fatalf("graceful shutdown failed: %v", err)
	}

	if time.Since(start) >= 3*time.Second {
		t.Fatal("shutdown exceeded its deadline")
	}

	select {
	case err := <-serveDone:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("expected http.ErrServerClosed, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server goroutine did not exit")
	}
}
