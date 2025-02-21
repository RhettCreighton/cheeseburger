package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServerGracefulShutdown(t *testing.T) {
	// Find an available port.
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Start the server with a simple handler.
	srv := &http.Server{
		Addr: fmt.Sprintf("localhost:%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate work.
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}),
	}

	// Start the server in a separate goroutine.
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	}()

	// Allow the server time to start.
	time.Sleep(50 * time.Millisecond)

	// Make a request to verify the server is running.
	go http.Get(fmt.Sprintf("http://localhost:%d/", port))

	// Initiate graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, srv.Shutdown(ctx))
}
