package Ports

import (
	"context"
	"gRPCbigapp/ClientService/Domain"
)

type CSOutboundPorts interface {
	SaveUser(ctx context.Context, user *Domain.User) error
	GetUser(ctx context.Context, userID string) (*Domain.User, error)
	UpdateUserPlan(ctx context.Context, userID string, role Domain.UserPlan) error
}
