package main

import (
	"gRPCbigapp/App/adapters/grpcAdapters"
	marketpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/markets"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

func main() {
	lis, _ := net.Listen("tcp", ":50052")

	service := grpcAdapters.NewMarketService()

	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	marketpb.RegisterSpotInstrumentServiceServer(server, service)

	grpc_prometheus.Register(server)

	server.Serve(lis)
}
