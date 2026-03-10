package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryPanicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Error(codes.Internal, "Упали с паникой")
			}
		}()
		return handler(ctx, req)
	}
}
