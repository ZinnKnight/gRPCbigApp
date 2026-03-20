package main

import (
	"fmt"
	"gRPCbigapp/app/adapters/grpcAdapters/marketAdapter"
	"gRPCbigapp/app/adapters/repo"
	"gRPCbigapp/app/interceptors/id_interceptor"
	"gRPCbigapp/app/interceptors/panic_interceptor"
	"gRPCbigapp/app/usecase"
	"gRPCbigapp/jaeger"
	logger2 "gRPCbigapp/logger"
	marketpb "gRPCbigapp/protofiles/gRPCbigapp/Protofiles/markets"
	redis2 "gRPCbigapp/redis"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		redisAddress = "localhost:6379"
	}

	redisDB := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	marketRepo := repo.NewRAMRepo()

	marketUC := usecase.NewMarketUseCaseImplementation(marketRepo)

	jaegerConfig := jaeger.JaegerConfiguration{
		EndpointCollector: os.Getenv("JAEGER_COLLECTOR_ENDPOINT"),
		ServiceName:       os.Getenv("JAEGER_SERVICE_NAME"),
	}
	if jaegerConfig.EndpointCollector == "" {
		jaegerConfig.EndpointCollector = "http://jaeger:14268"
	}
	if jaegerConfig.ServiceName == "" {
		jaegerConfig.ServiceName = "market-service"
	}

	grpcprometheus.EnableHandlingTimeHistogram()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			id_interceptor.RequestIdInterceptor,
			jaeger.JaegerInterceptor(jaegerConfig, logger),
			logger2.LoggerZapInterceptor(logger),
			panic_interceptor.UnaryPanicRecoveryInterceptor(logger),
			grpcprometheus.UnaryServerInterceptor,
			redis2.RedisCacheInterceptor(redisDB, 30*time.Second, logger),
		))

	marketGRPC := marketAdapter.NewMarketGrpcAdapter(marketUC)
	marketpb.RegisterSpotInstrumentServiceServer(server, marketGRPC)
	grpcprometheus.Register(server)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		logger.Info(fmt.Sprintf("Слушаем метрики на порте: %s", os.Getenv("PORT")))
		if err := http.ListenAndServe(":2113", mux); err != nil {
			logger.Error("Не можем достучатся до prometheus", zap.Error(err))
		}
	}()

	listening, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal("Не можем подключится", zap.Error(err))
	}
	fmt.Printf("MarketService слушает на: %s\n", os.Getenv("PORT"))
	if err := server.Serve(listening); err != nil {
		log.Fatal("Не можем читать", zap.Error(err))
	}
}
