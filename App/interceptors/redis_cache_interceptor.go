package interceptors

import (
	"context"
	"encoding/json"
	marketpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/markets"
	"time"

	"github.com/redis/go-redis"
	"google.golang.org/grpc"
)

// Я не могу разобратся что к чему в редисе больше недели и до сих пор туплю, я его душу

func RedisCacheInterceptor(rdc *redis.Client) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod != "market.SpotInstrumentService/ViewMarkets" {
			return handler(ctx, req)
		}

		key := "markets_cache"

		v, err := rdc.Get(key).Result()

		if err == nil {
			var response marketpb.ViewMarketsResponse

			err = json.Unmarshal([]byte(v), &response)
			if err != nil {
				return nil, err
			}
			return &response, nil
		}
		response, err := handler(ctx, req)

		if err == nil {
			cacheData, _ := json.Marshal(response)

			rdc.Set(key, cacheData, time.Minute)
		}
		return response, err
	}
}
