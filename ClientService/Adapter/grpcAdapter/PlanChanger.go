package grpcAdapter

import (
	"context"
	"gRPCbigapp/Shared/AuthShared/AuthCTX"
)

type PlanUpgradePreRequest interface {
	UpgradeAgree(ctx context.Context, user *AuthCTX.UserAuth) bool
}

type PlanChangePreRequestStub struct{}

func NewPlanChangePreRequestStub() PlanChangePreRequestStub {
	return PlanChangePreRequestStub{}
}

func (PlanChangePreRequestStub) UpgradeAgree(ctx context.Context, user *AuthCTX.UserAuth) bool {
	return true
}
