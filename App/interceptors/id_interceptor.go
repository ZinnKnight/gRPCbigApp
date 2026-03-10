package interceptors

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type ctxKey string

const RequestIDKey ctxKey = "x-request-id"

func XRequestIDInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	id := uuid.New().String()
	ctx = context.WithValue(ctx, RequestIDKey, id)
	return handler(ctx, req)
}
