package RateLimiter

import (
	"context"
	"fmt"
	"time"

	tracing "gRPCbigapp/Shared/Tracing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = tracing.Tracer("ratelimiter.redis")

// скрипт для редиса стянул с инета
var slidingWindowScript = redis.NewScript(
	`local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = tonumber(ARGV[4])

redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, now - window)
local count = redis.call("ZCARD", KEYS[1])

if count < limit then
redis.call('ZADD', KEYS[1], now, member)
redis.call('PEXPIRE', KEYS[1], window)
return {1, limit - count - 1, 0}
end 

local oldest = redis.call("ZRANGE", KEYS[1], 0, 0, 'WITHSCORES')
local retry = 0
if oldest[2] then
retry = (tonumber(oldest[2]) + window) - now
if retry < 0 then retry = 0 end 
end 
return {0, 0, retry}
`)

// просто маппер типов
func toInt(val interface{}) int64 {
	switch incomingVal := val.(type) {
	case int64:
		return incomingVal
	case int:
		return int64(incomingVal)
	default:
		return 0
	}
}

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

func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *Limiter {
	return &Limiter{
		rdb:           rdb,
		limit:         limit,
		slidingWindow: window,
	}
}

func (rl *Limiter) Allow(ctx context.Context, userID string) (*RateLimiterRes, error) {
	ctx, span := tracer.Start(ctx, "ratelimiter.Allow", tracing.KindClient)
	defer span.End()
	span.SetAttributes(tracing.RedisDB("EVAL")...)
	span.SetAttributes(attribute.String("user.id", userID))

	key := fmt.Sprintf("rate_limit:%s", userID)
	nowMs := time.Now().UnixMilli()

	windowMs := rl.slidingWindow.Milliseconds()

	member := fmt.Sprintf("%d - %s", nowMs, uuid.NewString())

	res, err := slidingWindowScript.Run(ctx, rl.rdb, []string{key}, nowMs, windowMs, rl.limit, member).Result()
	if err != nil {
		return nil, fmt.Errorf("ratelimiter: sliding window eval: %w", err)
	}

	vals, ok := res.([]interface{})
	if !ok || len(vals) != 3 {
		return nil, fmt.Errorf("ratelimiter: sliding window invalid result: %v", res)
	}

	allowed := toInt(vals[0]) == 1
	remaning := int(toInt(vals[1]))
	retryAfter := time.Duration(toInt(vals[2])) * time.Millisecond

	span.SetAttributes(
		attribute.Bool("ratelimit.allowed", allowed),
		attribute.Int("ratelimit.remaining", remaning),
	)
	if !allowed {
		span.AddEvent("rate_limit_exceeded")
	}

	return &RateLimiterRes{
		Allowed:    allowed,
		Remaining:  remaning,
		RetryAfter: retryAfter}, nil
}
