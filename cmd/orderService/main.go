package main

import (
	"context"
	"gRPCbigapp/App/adapters/grpcAdapters"
	"gRPCbigapp/App/adapters/postgra"
	"gRPCbigapp/App/interceptors"
	orderpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/order"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/postgres"
	}
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		logger.Fatal("Невозможно подключится к бд", zap.Error(err))
	}
	rds := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	repo := postgra.NewOrderRepoService(db)
	service := grpcAdapters.NewOrderService(repo)

	lis, _ := net.Listen("tcp", ":50051")

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryPanicRecoveryInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			interceptors.RedisCacheInterceptor(rds),
			//	grpc_prometheus.StreamServerInterceptor,
		),
	)

	orderpb.RegisterOrderServiceServer(server, service)

	grpc_prometheus.Register(server)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	logger.Info("Сервер поднят")

	server.Serve(lis)
}
