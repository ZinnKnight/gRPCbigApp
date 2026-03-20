package id_interceptor

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func TestRequestIdAdd(t *testing.T) {

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		id := ctx.Value(RequestID)

		if id == nil {
			t.Fatalf("x-request-id не добавлен: %v", id)
		}
		return "ok", nil
	}

	_, err := RequestIdInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("ошибка %v", err)
	}
}
