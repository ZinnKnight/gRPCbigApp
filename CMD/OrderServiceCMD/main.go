package main

import (
	"context"
	"fmt"
	clientGRPC "gRPCbigapp/App/ClientService/Adapter/grpcAdapter"
	clientPG "gRPCbigapp/App/ClientService/Adapter/postgresAdapter"
	clientUC "gRPCbigapp/App/ClientService/CSUseCase"
	orderPG "gRPCbigapp/App/OrderService/OSAdapters/OSPostgre"
	orderGRPC "gRPCbigapp/App/OrderService/OSAdapters/grpcAdapter"
	orderUC "gRPCbigapp/App/OrderService/OSUseCase"
	authAdapter "gRPCbigapp/App/Shared/Auth/AuthAdapter"
	authInterceptor "gRPCbigapp/App/Shared/Auth/AuthInterceptor"
	"gRPCbigapp/App/Shared/Config"
	logAdapter "gRPCbigapp/App/Shared/Logger/LoggerAdapter"
	"gRPCbigapp/App/Shared/Metrics"
	"gRPCbigapp/App/Shared/Outbox"
	"gRPCbigapp/App/Shared/PanicInterceptor"
	"gRPCbigapp/App/Shared/RateLimiter"
	redisClient "gRPCbigapp/App/Shared/Redis"
	"gRPCbigapp/App/Shared/Txmanager"
	orderPB "gRPCbigapp/Proto/Order"
	clientPB "gRPCbigapp/Proto/client"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := Config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v", err)
		os.Exit(1)
	}

	logger, err := logAdapter.NewZapLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v", err)
		os.Exit(1)
	}
	defer logger.Sync()

	poolCFG, err := pgxpool.ParseConfig(cfg.DatabaseURl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pgxpool.ParseConfig: %v", err)
		os.Exit(1)
	}
	poolCFG.MaxConns = 20
	poolCFG.MinConns = 2

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCFG)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pgxpool.New: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "pool.Ping: %v", err)
		os.Exit(1)
	}

	rdb, err := redisClient.NewRedisClient(ctx, cfg.RedisPassword, cfg.RedisAddr, cfg.RedisDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "redisClient.NewRedisClient: %v", err)
		os.Exit(1)
	}
	defer rdb.Close()

	txManag := Txmanager.NewTxManager(pool)
	outboxRepo := Outbox.NewRepository(pool)

	orderRepo := orderPG.NewOrderRepo(pool)
	orderUseCase := orderUC.NewOSUseCase(orderRepo, outboxRepo, txManag, logger)

	clientRepo := clientPG.NewUserRepo(pool)
	clientUseCase := clientUC.NewUserUseCase(clientRepo, outboxRepo, txManag, logger)

	limiter := RateLimiter.NewRateLimiter(rdb.Client, cfg.RateLimitPerMin, time.Minute)

	jwtService := authAdapter.NewJWTService([]byte(cfg.JWTSecretKey), 4*time.Hour)
	orderHandler := orderGRPC.NewOrderHandler(logger, orderUseCase)
	clientHandler := clientGRPC.NewUserhandler(clientUseCase, logger, jwtService)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			PanicInterceptor.PanicRecoveryInterceptor(logger),
			Metrics.UnaryServerInterceptor(),
			authInterceptor.AuthInterceptor([]byte(cfg.JWTSecretKey)),
			RateLimiter.UnaryServerInterceptor(limiter),
		),
	)

	orderPB.RegisterOrderServiceServer(grpcServer, orderHandler)
	clientPB.RegisterAuthServiceServer(grpcServer, clientHandler)

	go func() {
		if err := Metrics.StartMetricsServer(ctx, cfg.MetricsPort); err != nil {
			fmt.Fprintf(os.Stderr, "Metrics.StartMetricsServer: %v", err)
		}
	}()

	relay := Outbox.NewRelay(
		outboxRepo,
		&noopPublisher{},
		logger,
		time.Duration(cfg.OutBoxInterval)*time.Second,
		cfg.OutBoxButchSize,
	)
	go relay.Start(ctx)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		fmt.Fprintf(os.Stderr, "net.Listen: %v", err)
		os.Exit(1)
	}

	go func() {
		sigCH := make(chan os.Signal, 1)
		signal.Notify(sigCH, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCH
		fmt.Printf("Received signal: %v, shutting down...\n", sig)

		grpcServer.GracefulStop()
		cancel()
	}()

	fmt.Printf("order-service listening on grpc=:%d metrics=:%d\n", cfg.GRPCPort, cfg.MetricsPort)
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "grpcServer.Serve: %v", err)
		os.Exit(1)
	}
}

// MOCK for kafka or smth like that

type noopPublisher struct{}

func (p *noopPublisher) Publish(_ context.Context, event *Outbox.Event) error {
	fmt.Printf("[noop-publisher] event_type=%s aggregate_id=%s\n", event.EventType, event.AggregatorID)
	return nil
}
