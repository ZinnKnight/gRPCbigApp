package CSPorts

import (
	"context"
	"gRPCbigapp/App/ClientService/CSDomain"
)

type CSOutboundPorts interface {
	SaveUser(ctx context.Context, user *CSDomain.User) error
	GetUser(ctx context.Context, userID string) (*CSDomain.User, error)
	UpdateUserPlan(ctx context.Context, userID string, role CSDomain.UserPlan) error
	IsAdmin(ctx context.Context, userID string) (bool, error)
}
