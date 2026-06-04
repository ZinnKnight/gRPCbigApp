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
	RegisterUser(ctx context.Context, input RegisterUserInput) (*Domain.User, error)
	LoginUser(ctx context.Context, userName, UserPassword string) (*Domain.User, error)
	ChangeUserPlan(ctx context.Context, userName string, newPlan Domain.UserPlan) (*Domain.User, error)
}
