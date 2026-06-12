package Ports

import (
	"context"
	"gRPCbigapp/App/ClientService/Domain"
)

type RegisterUserInput struct {
	UserName     string
	UserPassword string
}

type UserInboundPort interface {
	UserRegistration(ctx context.Context, input RegisterUserInput) (*Domain.User, error)
	UserLogin(ctx context.Context, userName, UserPassword string) (*Domain.User, error)
	PlanChange(ctx context.Context, userName string, newPlan Domain.UserPlan) (*Domain.User, error)
}
