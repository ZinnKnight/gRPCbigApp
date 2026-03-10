package tests

import (
	"context"
	"gRPCbigapp/App/interceptors"
	"testing"

	"google.golang.org/grpc"
)

func TestRequestIdAdd(t *testing.T) {

	inter := interceptors.XRequestIDInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {

		id := ctx.Value(interceptors.RequestIDKey)

		if id == nil {
			t.Fatalf("x-request-id не добавлен: %v", id)
		}
		return "ok", nil
	}

	_, err := inter(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf(err)
	}
}
