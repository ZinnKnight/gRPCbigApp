package interceptors

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// я натрахался с тем как это ставить просто до нельзя

func LoggerZapInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Info("Запрос на конект grpcAdapters", zap.String("method", info.FullMethod))

		resp, err := handler(ctx, req)

		if err != nil {
			logger.Error("ошибка в grpcAdapters", zap.Error(err))
		}

		return resp, err
	}
}
