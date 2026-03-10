package main

import (
	"context"
	"gRPCbigapp/App/adapters/grpcAdapters"
	"gRPCbigapp/App/adapters/postgra"
	orderpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/order"
	"log"
	"net"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()

	db, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@postgres:5432/exchange")
	if err != nil {
		log.Fatal(err)
	}

	repo := postgra.NewOrderRepoService(db)
	service := grpcAdapters.NewOrderService(repo)

	lis, _ := net.Listen("tcp", ":50051")

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	orderpb.RegisterOrderServiceServer(server, service)

	grpc_prometheus.Register(server)

	logger.Info("Сервер поднят")

	server.Serve(lis)
}
