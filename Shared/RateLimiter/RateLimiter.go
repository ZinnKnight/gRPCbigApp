package RateLimiter

import (
	"context"
	"fmt"
	"time"

	tracing "gRPCbigapp/Shared/Tracing"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = tracing.Tracer("ratelimiter.redis")

type Limiter struct {
	rdb           *redis.Client
	limit         int
	slidingWindow time.Duration
}

type RateLimiterRes struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

func NewRateLimiter(rdb *redis.Client, limit int, slidingWindow time.Duration) *Limiter {
	return &Limiter{
		rdb:           rdb,
		limit:         limit,
		slidingWindow: slidingWindow,
	}
}

func (rl *Limiter) Allow(ctx context.Context, userID string) (*RateLimiterRes, error) {
	ctx, span := tracer.Start(ctx, "ratelimiter.Allow", tracing.KindClient)
	defer span.End()
	span.SetAttributes(tracing.RedisDB("INCR")...)
	span.SetAttributes(attribute.String("user.id", userID))

	key := fmt.Sprintf("rate_limit:%s", userID)

	count, err := rl.rdb.Incr(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("rateLimiterRes, failed to incr rate limit: %w", err)
	}

	if count == 1 {
		rl.rdb.Expire(ctx, key, rl.slidingWindow)
	}

	if count > int64(rl.limit) {
		ttl, _ := rl.rdb.TTL(ctx, key).Result()

		span.SetAttributes(
			attribute.Bool("ratelimit.allowed", false),
			attribute.Int64("ratelimit.count", count),
		)

		span.AddEvent("rate_limit_exceeded")

		return &RateLimiterRes{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: ttl,
		}, nil
	}

	span.SetAttributes(attribute.Bool("ratelimit.allowed", true),
		attribute.Int64("ratelimit.count", count),
	)

	return &RateLimiterRes{
		Allowed:   true,
		Remaining: rl.limit - int(count),
	}, nil
}
