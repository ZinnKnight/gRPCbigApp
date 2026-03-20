package panic_interceptor

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryPanicRecoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("востанавливаемся от паники",
					zap.String("метод", info.FullMethod),
					zap.Any("паника", r),
				)
				err = status.Errorf(codes.Internal, fmt.Sprintf("ошибка типа Internal: %v", r))
			}

		}()
		return handler(ctx, req)
	}
}
