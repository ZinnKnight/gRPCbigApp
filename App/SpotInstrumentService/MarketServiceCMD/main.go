package main

import (
	"context"
	"fmt"
	authInterceptor "gRPCbigapp/Shared/Auth/AuthInterceptor"
	"gRPCbigapp/Shared/Config"
	logAdapter "gRPCbigapp/Shared/Logger/LoggerAdapter"
	Metrics2 "gRPCbigapp/Shared/Metrics"
	"gRPCbigapp/Shared/PanicInterceptor"
	RateLimiter2 "gRPCbigapp/Shared/RateLimiter"
	redisClient "gRPCbigapp/Shared/Redis"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	marketPG "gRPCbigapp/App/SpotInstrumentService/Adapters/Postgres"
	marketGRPC "gRPCbigapp/App/SpotInstrumentService/Adapters/SISgrpcAdapter"
	marketUC "gRPCbigapp/App/SpotInstrumentService/SISUseCase"
	marketPB "gRPCbigapp/Proto/market"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := Config.LoadConfig()
	if err != nil {
		// Probably should do a panic, Tuzov do it his yt vid ab grpc, but idk what is propper way for "real-prod" way
		fmt.Println("Error loading config: %w", err)
		os.Exit(1)
	}

	logger, err := logAdapter.NewZapLogger()
	if err != nil {
		fmt.Println("Error initializing logger: %w", err)
		os.Exit(1)
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURl)
	if err != nil {
		fmt.Println("Error initializing postgres pool: %w", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Println("Error pinging postgres pool: %w", err)
		os.Exit(1)
	}

	rdb, err := redisClient.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		fmt.Println("Error initializing redis client: %w", err)
		os.Exit(1)
	}
	defer rdb.Close()

	marketRepo := marketPG.NewSISMarketRepo(pool)
	marketUseCase := marketUC.NewSISUseCase(marketRepo, logger)
	marketHandler := marketGRPC.NewSISgrpcHandler(marketUseCase, logger)

	limiter := RateLimiter2.NewRateLimiter(rdb.Client, cfg.RateLimitPerMin, time.Minute)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			PanicInterceptor.PanicRecoveryInterceptor(logger),
			Metrics2.UnaryServerInterceptor(),
			authInterceptor.AuthInterceptor([]byte(cfg.JWTSecretKey)),
			RateLimiter2.UnaryServerInterceptor(limiter),
		),
	)
	marketPB.RegisterSpotInstrumentServiceServer(grpcServer, marketHandler)

	go func() {
		if err := Metrics2.StartMetricsServer(ctx, cfg.MetricsPort); err != nil {
			fmt.Println("Error starting metrics server: %w", err)
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v", err)
		os.Exit(1)
	}

	go func() {
		sigCH := make(chan os.Signal, 1)
		signal.Notify(sigCH, syscall.SIGINT, syscall.SIGTERM)
		<-sigCH
		grpcServer.GracefulStop()
		cancel()
	}()

	fmt.Printf("market server listening on grpc =: %d, metrics = :%d\n", cfg.GRPCPort, cfg.MetricsPort)
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "failed to serve: %v", err)
		os.Exit(1)
	}
}
