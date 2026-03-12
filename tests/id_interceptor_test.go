package tests

import (
	"context"
	"gRPCbigapp/App/interceptors"
	"testing"

	"google.golang.org/grpc"
)

func TestRequestIdAdd(t *testing.T) {

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		id := ctx.Value(interceptors.RequestID)

		if id == nil {
			t.Fatalf("x-request-id не добавлен: %v", id)
		}
		return "ok", nil
	}

	_, err := interceptors.RequestIdInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("ошибка %v", err)
	}
}
