package jaeger

import (
	"context"
	"fmt"
	"gRPCbigapp/app/interceptors/id_interceptor"
	"io"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type JaegerConfiguration struct {
	EndpointCollector string
	ServiceName       string
}

func infoSenderJaeger(conf JaegerConfiguration, method, requestId string, logger *zap.Logger) {
	if conf.EndpointCollector == "" {
		return
	}
	url := fmt.Sprintf("%s/app/traces", conf.EndpointCollector)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		// использовал Debug что бы можно было легко траблшутить, можно будет поменять на Warn при необходимости
		logger.Debug("Ошибка при создании запроса на отправку", zap.Error(err))
		return
	}
	req.Header.Set("x-request-id", requestId)
	req.Header.Set("x-service-name", conf.ServiceName)
	req.Header.Set("метод", method)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Debug("Jaeger метод send умер", zap.Error(err))
		return
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()
}

func JaegerInterceptor(conf JaegerConfiguration, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		requestID := ""
		if id, ok := ctx.Value(id_interceptor.RequestID).(string); ok {
			requestID = id
		}
		if metaD, ok := metadata.FromIncomingContext(ctx); ok {
			if val := metaD.Get("x-request-id"); len(val) > 0 && requestID == "" {
				requestID = val[0]
			}
		}

		logger.Info("Jaeger interceptor",
			zap.String("Сервис:", conf.ServiceName),
			zap.String("Метод:", info.FullMethod),
		)
		go infoSenderJaeger(conf, info.FullMethod, requestID, logger)
		return handler(ctx, req)
	}
}
