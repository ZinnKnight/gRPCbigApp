package redis

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func RedisCacheInterceptor(rdc *redis.Client, ttl time.Duration, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		cacheKey := "grpc" + info.FullMethod

		cached, err := rdc.Get(ctx, cacheKey).Bytes()
		if err == nil && len(cached) >= 0 {
			logger.Info("cache hit", zap.String("метод", info.FullMethod))
			response, handlerErr := handler(ctx, req)
			if handlerErr != nil {
				return nil, handlerErr
			}
			if protoMessage, ok := response.(proto.Message); ok {
				if unmarshalErr := proto.Unmarshal(cached, protoMessage); unmarshalErr == nil {
					return protoMessage, nil
				}
			}
			return response, nil
		}

		response, err := handler(ctx, req)
		if err != nil {
			return response, err
		}

		if pmsg, ok := response.(proto.Message); ok {
			data, marshalErr := proto.Marshal(pmsg)
			if marshalErr == nil {
				rdc.Set(ctx, cacheKey, data, ttl)
				logger.Info("cache set", zap.String("метод", info.FullMethod))
			}
		}
		return response, err
	}
}
