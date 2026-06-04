package ValidationIntercepter

import (
	"context"
	"gRPCbigapp/Shared/ErrorInterceptor"

	"google.golang.org/grpc"
)

type allErrorsValidation interface {
	ValidateAll() error
}

// fallback
type legecyValidator interface {
	Validate() error
}

func executeValidation(req interface{}) error {
	if val, ok := req.(allErrorsValidation); ok {
		return val.ValidateAll()
	}
	if val, ok := req.(legecyValidator); ok {
		return val.Validate()
	}
	return nil
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		if err := executeValidation(req); err != nil {
			return nil, ErrorInterceptor.NewError(ErrorInterceptor.Invalid,
				"Некорректные данные, невозможно обработать запрос "+err.Error(), err)
		}
		return handler(ctx, req)
	}
}
