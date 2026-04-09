package grpcAdapter

import (
	"context"
	"errors"
	"gRPCbigapp/App/ClientService/CSDomain"
	"gRPCbigapp/App/ClientService/CSPorts"
	"gRPCbigapp/App/Shared/Auth/AuthAdapter"
	"gRPCbigapp/App/Shared/Auth/AuthCTX"
	"gRPCbigapp/App/Shared/Logger/LoggerPorts"
	clientpb "gRPCbigapp/Proto/client"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	clientpb.UnimplementedAuthServiceServer
	useCase CSPorts.UserInboundPort
	jwt     *AuthAdapter.JWTService
	logger  LoggerPorts.Logger
}

func NewUserhandler(uc CSPorts.UserInboundPort, log LoggerPorts.Logger, j *AuthAdapter.JWTService) *UserHandler {
	return &UserHandler{
		useCase: uc,
		jwt:     j,
		logger:  log,
	}
}

func UserErrorsMapper(err error) error {
	switch {
	case errors.Is(err, CSDomain.ErrEmptyName), errors.Is(err, CSDomain.ErrEmptyPassword):
		return status.Errorf(codes.InvalidArgument, err.Error())
	case errors.Is(err, CSDomain.ErrUserAlreadyExists):
		return status.Errorf(codes.AlreadyExists, err.Error())
	case errors.Is(err, CSDomain.ErrUserNotFound):
		return status.Errorf(codes.NotFound, err.Error())
	default:
		return status.Errorf(codes.Internal, err.Error())
	}
}

func (uh *UserHandler) UserRegistration(ctx context.Context, req *clientpb.RegisterRequest) (*clientpb.AuthResponse, error) {
	rui := CSPorts.RegisterUserInput{
		UserName:     req.GetUserName(),
		UserPassword: req.GetUserPassword(),
	}

	user, err := uh.useCase.RegisterUser(ctx, rui)
	if err != nil {
		uh.logger.LogError("grpc, failed to register user",
			LoggerPorts.Fieled{Key: "error", Value: err.Error()})
		return nil, UserErrorsMapper(err)
	}

	token, err := uh.jwt.GenerateToken(user.UserID, user.UserName, string(user.UserRole))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}
	return &clientpb.AuthResponse{Token: token}, nil
}

func (uh *UserHandler) UserLogin(ctx context.Context, req *clientpb.LoginRequest) (*clientpb.AuthResponse, error) {
	user, err := uh.useCase.LoginUser(ctx, req.UserId)
	if err != nil {
		return nil, UserErrorsMapper(err)
	}

	token, err := uh.jwt.GenerateToken(user.UserID, user.UserName, string(user.UserRole))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}
	return &clientpb.AuthResponse{Token: token}, nil
}

func (uh *UserHandler) IsAdmin(ctx context.Context, req *clientpb.IsAdminRequest) (*clientpb.IsAdminResponse, error) {
	isAdmin, err := uh.useCase.IsAdmin(ctx, req.UserId)
	if err != nil {
		return nil, UserErrorsMapper(err)
	}
	return &clientpb.IsAdminResponse{UserIsAdmin: isAdmin}, nil
}

func (uh *UserHandler) ChangeUserPlan(ctx context.Context, req *clientpb.PlanChangeRequest) (*clientpb.PlanChangeResponse, error) {
	plan, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	err := uh.useCase.ChangeUserPlan(ctx, plan.UserID, CSDomain.Pro)
	if err != nil {
		return nil, UserErrorsMapper(err)
	}

	token, err := uh.jwt.GenerateToken(plan.UserID, plan.UserName, string(plan.UserPlan))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}
	return &clientpb.PlanChangeResponse{Token: token}, nil
}
