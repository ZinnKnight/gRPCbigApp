package Quota

import (
	"context"
	"fmt"
	"gRPCbigapp/Shared/Policy"
	RateLimiter2 "gRPCbigapp/Shared/RateLimiter"
	"time"
)

type Decision struct {
	Allowed    bool
	RetryAfter time.Duration
}

// перенести в метадату

type Enforced struct {
	provider Policy.Provider
	limiter  *RateLimiter2.Limiter
}

func NewEnforced(provider Policy.Provider, limiter *RateLimiter2.Limiter) *Enforced {
	return &Enforced{
		provider: provider,
		limiter:  limiter,
	}
}

func (e *Enforced) Check(ctx context.Context, plan string, action Policy.Action, subject string) (Decision, error) {
	rule := e.provider.RuleFor(plan, action)

	if rule.Limit <= 0 {
		return Decision{Allowed: true}, nil
	}

	key := fmt.Sprintf("%s_%s", subject, plan)
	res, err := e.limiter.AllowKey(ctx, key, rule.Limit, rule.Window)
	if err != nil {
		return Decision{}, fmt.Errorf("quota, check: %s, for: %s: %w", action, subject, err)
	}
	return Decision{Allowed: true, RetryAfter: res.RetryAfter}, nil
}
