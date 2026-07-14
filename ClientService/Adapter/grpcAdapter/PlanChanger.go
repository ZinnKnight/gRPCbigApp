package grpcAdapter

import (
	"context"
	"gRPCbigapp/ClientService/Auth/AuthCTX"
)

type PlanUpgradePreRequest interface {
	UpgradeAgree(ctx context.Context, user *AuthCTX.UserAuth) bool
}

// пока это всё - апгрейднутый мок, что был до этого
// причина - я ещё не знаю какие условия для выполнения смены будут.
// На данный момент - захотел - поменял

type PlanChangePreRequestStub struct{}

func NewPlanChangePreRequestStub() PlanChangePreRequestStub {
	return PlanChangePreRequestStub{}
}

func (PlanChangePreRequestStub) UpgradeAgree(ctx context.Context, user *AuthCTX.UserAuth) bool {
	return true
}
