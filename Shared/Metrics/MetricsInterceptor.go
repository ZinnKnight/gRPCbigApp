package Metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()

		code := status.Code(err)

		GRPCRequestTotal.WithLabelValues(info.FullMethod, code.String()).Inc()

		GRPCRequestTotal.WithLabelValues(info.FullMethod, code.String()).Add(duration)

		return resp, err
	}
}
