package main

import (
	"context"
	"fmt"
	orderApp "gRPCbigapp/App/OrderService/App"
	"gRPCbigapp/Shared/Config"
	"gRPCbigapp/Shared/Logger/LoggerAdapter"
	"os"
	"os/signal"
	"syscall"
)

func run() error {
	cfg, err := Config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %v", err)
	}
	logger, err := LoggerAdapter.NewZapLogger()
	if err != nil {
		return fmt.Errorf("new logger: %v", err)
	}
	defer logger.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := orderApp.New(ctx, cfg, logger)
	if err != nil {
		return err
	}
	return application.Run(ctx)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "order-service: %v/n", err)
		os.Exit(1)
	}
}
