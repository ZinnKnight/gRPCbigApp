package interceptors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

// Я не могу нормально разобратся что к чему в редисе больше недели и до сих пор туплю.

func RedisCacheInterceptor(rdc *redis.Client) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		reqBytes, err := json.Marshal(req)
		if err != nil {
			log.Printf("Не удалось сериализовать данные кэша в Redis: %v", err)
			return handler(ctx, req)
		}

		key := fmt.Sprintf("cache %s:%s", info.FullMethod, reqBytes)

		data, err := rdc.Get(ctx, key).Bytes()
		if err == nil {
			// тут бутылочное горлышко будет если правильно понимаю, и что бы поправить, возможно, можно заюзать штуку что мы делали
			// в одном из заданий - а именно мы сами кастили json через pool вроде, и потом прокидывали обратно
			msg := reflect.New(
				reflect.TypeOf(req).Elem()).Interface().(proto.Message)

			if unmarshalErr := proto.Unmarshal(data, msg); unmarshalErr == nil {
				return msg, nil
			}
		}

		resp, err := handler(ctx, req)

		if err != nil {
			return resp, err
		}

		if msg, ok := resp.(proto.Message); ok {
			bytes, marshalErr := proto.Marshal(msg)
			if marshalErr == nil {
				rdc.Set(ctx, key, bytes, time.Minute)
			}
		}
		return resp, err
	}
}
