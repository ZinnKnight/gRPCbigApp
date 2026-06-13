package Metrics

import (
	"context"
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
		Addr:    fmt.Sprintf("port :%v", port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error in metrics server: %w", err)
	}
	return nil
}
