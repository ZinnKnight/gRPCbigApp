package Tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func gRPCDial(ctx context.Context, target string, extra ...grpc.DialOption) (*grpc.ClientConn, error) {
	if target == "" {
		return nil, fmt.Errorf("tracing: empty grpc address")
	}

	opts := []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	opts = append(opts, extra...)

	_, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return grpc.NewClient(target, opts...)
}
