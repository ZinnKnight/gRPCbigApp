package RateLimiter

import (
	"context"
	"fmt"
	"gRPCbigapp/Shared/Auth/AuthCTX"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(limit *Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		user, ok := AuthCTX.GetUser(ctx)
		if !ok {
			return handler(ctx, req)
		}

		res, err := limit.Allow(ctx, user.UserID)
		if err != nil {
			return handler(ctx, req)
		}

		if !res.Allowed {
			return nil, status.Error(codes.ResourceExhausted, fmt.Sprintf("Rate limit exceeded, retry after: %s", res.RetryAfter))
		}
		return handler(ctx, req)
	}
}
