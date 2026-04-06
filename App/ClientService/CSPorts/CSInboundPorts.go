package CSInboundPorts

import (
	"context"
	"gRPCbigapp/App/ClientService/CSDomain"
	"gRPCbigapp/App/ClientService/CSUseCase"
	"gRPCbigapp/App/Shared/Auth/AuthUseCase"
	"gRPCbigapp/App/Shared/Logger/LoggerPorts"
	clientpb "gRPCbigapp/Proto/client"

	"github.com/google/uuid"
)

type CSInboundHandler struct {
	inbH   *CSUseCase.UserService
	logger LoggerPorts.Logger
}

func NewCSLoggerService(log LoggerPorts.Logger, inbH *CSUseCase.UserService) *CSInboundHandler {
	return &CSInboundHandler{inbH: inbH, logger: log}
}

func (h *CSInboundHandler) CreateUser(ctx context.Context, req *clientpb.RegisterRequest) (*clientpb.AuthResponse, error) {
	user := &CSDomain.User{
		UserID:       uuid.New().String(),
		UserName:     req.UserName,
		UserPassword: req.UserPassword,
		UserRole:     "FREE_PLAN_USER",
	}

	err := h.inbH.CreateUser(ctx, user)
	if err != nil {

		// TODO ROLLBACK or some type of that

		h.logger.LogError("Error in CreteUser method",
			LoggerPorts.Fieled{Key: "user_id", Value: user.UserID},
			LoggerPorts.Fieled{Key: "Error", Value: err.Error()})
		return nil, err
	}
	return &clientpb.AuthResponse{
		// TODO TOKEN
	}, nil
}

func (h *CSInboundHandler) LoginUser(ctx context.Context, req *clientpb.LoginRequest) (*clientpb.AuthResponse, error) {
	user, err := h.inbH.LoginUser(ctx, req.UserId)
	if err != nil {

		// TODO ROLLBACK or some type of that

		h.logger.LogError("Error in LoginUser method",
			LoggerPorts.Fieled{Key: "user_id", Value: user},
			LoggerPorts.Fieled{Key: "Error", Value: err.Error()})
		return nil, err
	}
	return &clientpb.AuthResponse{
		// TODO TOKEN JWT
	}, nil
}

func (h *CSInboundHandler) IsAdmin(ctx context.Context, req *clientpb.IsAdminRequest) (*clientpb.IsAdminResponse, error) {
	// TODO probably should recheck how todo it properly
	_, err := h.inbH.IsAdmin(ctx, req.UserId)
	if err != nil {

		// TODO ROLLBACK or some type of that

		h.logger.LogError("Error in IsAdmin method",
			LoggerPorts.Fieled{Key: "user_id", Value: req.UserId},
			LoggerPorts.Fieled{Key: "Error", Value: err.Error()})
		return nil, err
	}
	return &clientpb.IsAdminResponse{
		UserIsAdmin: true,
	}, nil
}

func (h *CSInboundHandler) UpdateUserPlan(ctx context.Context, req *clientpb.PlanChangeRequest) (*clientpb.PlanChangeResponse, error) {
	user := AuthUseCase.GetUserFromContext(ctx)
	updateUser := &CSDomain.User{
		UserID:   user.UserID,
		UserName: user.UserName,
		UserRole: "PRO_PLAN_USER",
	}
	err := h.inbH.ChangeUserPlan(ctx, updateUser)
	if err != nil {

		// TODO ROLLBACK or some type of that

		h.logger.LogError("Error in UpdateUserPlan method",
			LoggerPorts.Fieled{Key: "user_id", Value: user.UserID},
			LoggerPorts.Fieled{Key: "Error", Value: err.Error()})
		return nil, err
	}
	return &clientpb.PlanChangeResponse{
		// TODO idk how to mountain it properly yet
	}, nil
}
