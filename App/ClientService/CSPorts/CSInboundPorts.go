package CSPorts

import (
	"context"
	"gRPCbigapp/App/ClientService/CSDomain"
)

// эту кострукцию добавил т.к в дальнейшем данноеобращение будет использоватся несколько раз
// оверхед на упрощение в цену скорости

type RegisterUserInput struct {
	UserName     string
	UserPassword string
}

type UserInboundPort interface {
	RegisterUser(ctx context.Context, input RegisterUserInput) (*CSDomain.User, error)
	LoginUser(ctx context.Context, userID string) (*CSDomain.User, error)
	IsAdmin(ctx context.Context, userID string) (bool, error)
	ChangeUserPlan(ctx context.Context, userID string, newPlan CSDomain.UserPlan) error
}
