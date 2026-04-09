package RateLimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

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
		return &RateLimiterRes{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: ttl,
		}, nil
	}
	return &RateLimiterRes{
		Allowed:   true,
		Remaining: rl.limit - int(count),
	}, nil
}
