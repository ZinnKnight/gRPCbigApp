package id_interceptor

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ctxKey string

const RequestID ctxKey = "x-request-id"

func RequestIdInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	requestID := ""

	if metaD, ok := metadata.FromIncomingContext(ctx); ok {
		if val := metaD.Get(string(RequestID)); len(val) > 0 {
			requestID = val[0]
		}
	}
	if requestID == "" {
		requestID = uuid.New().String()
	}
	context.WithValue(ctx, RequestID, requestID)
	return handler(ctx, req)
}

func XRequestIdInterceptor(ctx context.Context, method string, requirement, reply interface{}, clientContract *grpc.ClientConn, invoke grpc.UnaryInvoker, options ...grpc.CallOption) error {
	requestID := uuid.New().String()
	if id, ok := ctx.Value(RequestID).(string); ok && id != "" {
		requestID = id
	}
	ctx = metadata.AppendToOutgoingContext(ctx, string(RequestID), requestID)
	return invoke(ctx, method, requirement, reply, clientContract, options...)
}
