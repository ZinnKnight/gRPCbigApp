package Metrics

import (
	"context"
	"gRPCbigapp/Shared/Metrics/MetricsPort"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(rec MetricsPort.MetricsRecord) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		resp, err := handler(ctx, req)

		rec.IncRequest(info.FullMethod, status.Code(err).String())

		return resp, err
	}
}
