package grpcAdapter

import "gRPCbigapp/Shared/ErrorInterceptor"

var ErrUnauthenticated = ErrorInterceptor.GRPCConnector(ErrorInterceptor.NewError(ErrorInterceptor.Unauthenticated, "Требуется авторизация", nil))
