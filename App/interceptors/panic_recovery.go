package interceptors

import (
	"context"
	"fmt"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//grpc-ecosystem/go-grpc-middleware
// нашёл что уже есть готовый итерцептор для паник уже после того как рачком-бочком написал свой
// для обучения мб норм, но в дальнейшем буду обращаться к этому

func UnaryPanicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Error(codes.Internal, fmt.Sprintf("Eternal server error: %s", info.FullMethod))
				fmt.Printf("Упали с паникой: %v, \n в методе: %s, \n в стаке: %s", r, info.FullMethod, debug.Stack())
			}

		}()
		return handler(ctx, req)
	}
}
