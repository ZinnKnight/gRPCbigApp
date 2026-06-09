package main

import (
	"context"
	"fmt"
	marketapp "gRPCbigapp/App/SpotInstrumentService/App"
	"gRPCbigapp/Shared/Config"
	"gRPCbigapp/Shared/Logger/LoggerAdapter"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "market service: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := Config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	logger, err := LoggerAdapter.NewZapLogger()
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer logger.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	applications, err := marketapp.New(ctx, cfg, logger)
	if err != nil {
		return err
	}

	return applications.Run(ctx)
}
