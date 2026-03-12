package interceptors

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ctxKey string

const RequestID ctxKey = "request-id"

func RequestIdInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	metaD, ok := metadata.FromIncomingContext(ctx)

	var requestId string

	if ok {
		ids := metaD.Get("x-request-id")
		if len(ids) > 0 {
			requestId = ids[0]
		}
	}

	if requestId == "" {
		requestId = uuid.New().String()
	}

	ctx = context.WithValue(ctx, RequestID, requestId)
	return handler(ctx, req)
}
