package CSUseCase

import "C"
import (
	"context"
	"gRPCbigapp/ClientService/CSDomain"
	"gRPCbigapp/ClientService/CSPorts/CSOutboundPorts"
)

type UserService struct {
	repo CSOutboundPorts.CSOutboundPorts
}

func NewUserService(r CSOutboundPorts.CSOutboundPorts) *UserService {
	return &UserService{repo: r}
}

func (us *UserService) CreateUser(ctx context.Context, user *CSDomain.User) error {
	return us.repo.SaveUserInDB(ctx, user)
}

func (us *UserService) LoginUser(ctx context.Context, userId string) (*CSDomain.User, error) {
	return us.repo.GetUserFromDB(ctx, userId)
}

func (us *UserService) IsAdmin(ctx context.Context, admin string) (*CSDomain.Admin, error) {
	return us.repo.CheckIsAdmin(ctx, admin)
}

func (us *UserService) ChangeUserPlan(ctx context.Context, user *CSDomain.User) error {
	return us.repo.UpdateUserPlan(ctx, user)
}
