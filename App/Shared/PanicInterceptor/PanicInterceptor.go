package PanicInterceptor

import (
	"context"
	"gRPCbigapp/App/Shared/Logger/LoggerPorts"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PanicRecoveryInterceptor(logger LoggerPorts.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (response interface{}, err error) {

		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())

				logger.LogError("Panic was prevented in grpc handler",
					LoggerPorts.Fieled{Key: "method", Value: info.FullMethod},
					LoggerPorts.Fieled{Key: "stack", Value: stack},
					LoggerPorts.Fieled{Key: "panic", Value: r},
				)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}
