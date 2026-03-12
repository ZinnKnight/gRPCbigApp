package main

import (
	"gRPCbigapp/App/adapters/grpcAdapters"
	"gRPCbigapp/App/interceptors"
	marketpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/markets"
	"log"
	"net"
	"net/http"

	grpcrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, _ := zap.NewProduction()

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Не удалось подключится: %v", err)
	}

	rds := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	service := grpcAdapters.NewMarketService()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryPanicRecoveryInterceptor(),
			grpcrometheus.UnaryServerInterceptor,
			interceptors.RedisCacheInterceptor(rds),
			interceptors.LoggerZapInterceptor(logger),
			interceptors.RequestIdInterceptor,
		),
	)

	marketpb.RegisterSpotInstrumentServiceServer(server, service)

	grpcrometheus.Register(server)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2113", nil)
	}()

	server.Serve(lis)
}
