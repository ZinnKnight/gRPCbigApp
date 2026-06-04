package CSPorts

import (
	"context"
	"gRPCbigapp/App/ClientService/CSDomain"
)

type RegisterUserInput struct {
	UserName     string
	UserPassword string
}

type UserInboundPort interface {
	RegisterUser(ctx context.Context, input RegisterUserInput) (*CSDomain.User, error)
	LoginUser(ctx context.Context, userName, UserPassword string) (*CSDomain.User, error)
	ChangeUserPlan(ctx context.Context, userName string, newPlan CSDomain.UserPlan) (*CSDomain.User, error)
}
