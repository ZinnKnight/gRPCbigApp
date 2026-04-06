package CSOutboundPorts

import (
	"context"
	"gRPCbigapp/App/ClientService/CSDomain"
)

type CSOutboundPorts interface {
	SaveUserInDB(ctx context.Context, user *CSDomain.User) error
	// send data ab user in db, when collecting it

	GetUserFromDB(ctx context.Context, userID string) (*CSDomain.User, error)
	// load a data ab user after login

	UpdateUserPlan(ctx context.Context, user *CSDomain.User) error
	// send data ab updating user role, when he decided change it

	CheckIsAdmin(ctx context.Context, admin string) (*CSDomain.Admin, error)
}
