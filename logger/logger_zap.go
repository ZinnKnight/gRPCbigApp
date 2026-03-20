package logger

import (
	"context"
	"gRPCbigapp/app/interceptors/id_interceptor"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LoggerZapInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		requestID := ""
		if id, ok := ctx.Value(id_interceptor.RequestID).(string); ok {
			requestID = id
		}
		logger.Info("gRPC запрос запущен",
			zap.String("метод", info.FullMethod),
			zap.String("x-request-id", requestID),
		)

		resp, err := handler(ctx, req)

		logger.Info("gRPC запрос выполнен",
			zap.String("метод", info.FullMethod),
			zap.String("x-request-id", requestID),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return resp, err
	}
}
