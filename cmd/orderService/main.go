package main

import (
	"context"
	"fmt"
	"gRPCbigapp/app/adapters/grpcAdapters/acsessAdapter"
	"gRPCbigapp/app/adapters/grpcAdapters/orderAdapter"
	"gRPCbigapp/app/interceptors/id_interceptor"
	"gRPCbigapp/app/interceptors/panic_interceptor"
	"gRPCbigapp/app/postgra"
	"gRPCbigapp/app/usecase"
	"gRPCbigapp/jaeger"
	logger2 "gRPCbigapp/logger"
	orderpb "gRPCbigapp/protofiles/gRPCbigapp/Protofiles/order"
	"log"
	"net"
	"net/http"
	"os"

	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/gRPCbigApp"
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal("Не может подключится к базе", zap.Error(err))
	}
	defer pool.Close()

	repository := postgra.NewOrderRepository(pool)
	if err := repository.DatabaseShemeInitiation(context.Background()); err != nil {
		log.Fatal("Не смогли создать разметку бд", zap.Error(err))
	}

	marketAddress := os.Getenv("MARKET_ADDRESS")
	if marketAddress == "" {
		marketAddress = "market-service:50052"
	}
	marketConnection, err := grpc.NewClient(marketAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(id_interceptor.XRequestIdInterceptor))
	if err != nil {
		log.Fatal("Не можем подключить market-service", zap.Error(err))
	}
	defer marketConnection.Close()

	marketCk := acsessAdapter.NewGRPCAccessAdapter(marketConnection)

	orderUC := usecase.NewOrderUseCaseImplementation(repository, marketCk)

	jaegerCfg := jaeger.JaegerConfiguration{
		EndpointCollector: os.Getenv("JAEGER_ENDPOINT_COLLECTOR"),
		ServiceName:       os.Getenv("JAEGER_SERVICE_NAME"),
	}
	if jaegerCfg.EndpointCollector == "" {
		jaegerCfg.EndpointCollector = "http://jaeger:14268"
	}
	if jaegerCfg.ServiceName == "" {
		jaegerCfg.ServiceName = "order-service"
	}

	grpcprometheus.EnableHandlingTimeHistogram()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			id_interceptor.RequestIdInterceptor,
			jaeger.JaegerInterceptor(jaegerCfg, logger),
			logger2.LoggerZapInterceptor(logger),
			panic_interceptor.UnaryPanicRecoveryInterceptor(logger),
			grpcprometheus.UnaryServerInterceptor,
		))

	orderServ := orderAdapter.NewOrderGrpcAdapter(orderUC)
	orderpb.RegisterOrderServiceServer(server, orderServ)
	grpcprometheus.Register(server)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		logger.Info("слушаем метрики prometheus :2112")
		if err := http.ListenAndServe(":2112", mux); err != nil {
			logger.Error("не можем достучатся до сервиса prometheus ", zap.Error(err))
		}
	}()

	listening, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("не можем подключится ", zap.Error(err))
	}
	fmt.Printf("Слушаем метрики на :50051")
	if err := server.Serve(listening); err != nil {
		log.Fatal("не можем читать", zap.Error(err))
	}
}
