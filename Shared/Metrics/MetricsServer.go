package Metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

func StartMetricsServer(ctx context.Context, port int, handler http.Handler) error {
	mux := http.NewServeMux()

	mux.Handle("/metrics", handler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error in metrics server: %w", err)
	}
	return nil
}
