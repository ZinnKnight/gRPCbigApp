package Redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}

type RedisClient struct {
	*redis.Client
}

func NewRedisClient(ctx context.Context, config Config) (*RedisClient, error) {
	options := &redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	}
	if config.PoolSize > 0 {
		options.PoolSize = config.PoolSize
	}
	if config.MinIdleConns > 0 {
		options.MinIdleConns = config.MinIdleConns
	}

	rdb := redis.NewClient(options)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping to redis: %w", err)
	}
	return &RedisClient{Client: rdb}, nil

}

func (rc *RedisClient) Close() error {
	return rc.Client.Close()
}
