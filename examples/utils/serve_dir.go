package utils

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// ServeDir starts an HTTP file server for the given local directory on addr.
// Example: ServeDir(":8080", "./public")
func ServeDir(ctx context.Context, addr string, dir string) error {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(dir)))

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Channel to capture server startup / runtime errors
	errCh := make(chan error, 1)

	// Start the server
	go func() {

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():

		// Give active connections time to finish
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		return server.Shutdown(shutdownCtx)

	case err := <-errCh:
		return err
	}
}
