package Redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
}

func NewRedisClient(ctx context.Context, addres, password string, redisDB int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addres,
		Password: password,
		DB:       redisDB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis, ping was failed: %w", err)
	}
	return &RedisClient{Client: rdb}, nil
}

func (rc *RedisClient) Close() error {
	return rc.Client.Close()
}
