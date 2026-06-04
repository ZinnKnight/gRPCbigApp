package ErrorInterceptor

import (
	"context"
	"gRPCbigapp/Shared/Logger/LoggerPorts"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(logger LoggerPorts.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		if IsHiden(err) && logger != nil {
			logger.LogError("rpc handler error",
				LoggerPorts.Field{Key: "method", Value: info.FullMethod},
				LoggerPorts.Field{Key: "error", Value: err.Error()},
			)
		}
		return resp, GRPCConnector(err)
	}
}
