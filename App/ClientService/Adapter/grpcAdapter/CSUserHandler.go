package grpcAdapter

import (
	"context"
	"errors"
	"gRPCbigapp/App/ClientService/CSDomain"
	"gRPCbigapp/App/ClientService/CSPorts"
	"gRPCbigapp/Proto/protoPB/clientPB"
	"gRPCbigapp/Shared/Auth/AuthAdapter"
	"gRPCbigapp/Shared/Auth/AuthCTX"
	"gRPCbigapp/Shared/Logger/LoggerPorts"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	clientPB.UnimplementedAuthServiceServer
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

// временный мок для смены роли самим юзером. Позже добавлю рестрикшены, сейчас мок на автоматический запрос-выдачу

func (uh *UserHandler) planUserChangeActionMock(_ context.Context, _ *AuthCTX.UserAuth) bool {
	return true
}

func UserErrorsMapper(err error) error {
	switch {
	case errors.Is(err, CSDomain.ErrEmptyName), errors.Is(err, CSDomain.ErrEmptyPassword):
		return status.Errorf(codes.InvalidArgument, err.Error())
	case errors.Is(err, CSDomain.ErrIncorrectCredentials):
		return status.Errorf(codes.InvalidArgument, err.Error())
	case errors.Is(err, CSDomain.ErrUserAlreadyExists):
		return status.Errorf(codes.AlreadyExists, err.Error())
	case errors.Is(err, CSDomain.ErrUserNotFound):
		return status.Errorf(codes.NotFound, err.Error())
	default:
		return status.Errorf(codes.Internal, err.Error())
	}
}

// маппинг ролей + чек на присваивание

func planToUser(plan CSDomain.UserPlan) clientPB.Roles {
	if val, ok := clientPB.Roles_value[string(plan)]; ok {
		return clientPB.Roles(val)
	}
	return clientPB.Roles_UNAUTHORISED_USER
}

func roleToUser(role clientPB.Roles) CSDomain.UserPlan {
	return CSDomain.UserPlan(role.String())
}

func (uh *UserHandler) UserRegistration(ctx context.Context, req *clientPB.RegisterRequest) (*clientPB.AuthResponse, error) {
	rui := CSPorts.RegisterUserInput{
		UserName:     req.GetUserName(),
		UserPassword: req.GetUserPassword(),
	}

	user, err := uh.useCase.RegisterUser(ctx, rui)
	if err != nil {
		uh.logger.LogError("grpc, failed to register user",
			LoggerPorts.Field{Key: "error", Value: err.Error()})
		return nil, UserErrorsMapper(err)
	}

	token, err := uh.jwt.GenerateToken(user.UserID, user.UserName, string(user.UserRole))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}
	return &clientPB.AuthResponse{Token: token}, nil
}

func (uh *UserHandler) UserLogin(ctx context.Context, req *clientPB.LoginRequest) (*clientPB.AuthResponse, error) {
	user, err := uh.useCase.LoginUser(ctx, req.UserName)
	if err != nil {
		return nil, UserErrorsMapper(err)
	}

	token, err := uh.jwt.GenerateToken(user.UserID, user.UserName, string(user.UserRole))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}
	return &clientPB.AuthResponse{Token: token}, nil
}

func (uh *UserHandler) IsAdmin(ctx context.Context, req *clientPB.IsAdminRequest) (*clientPB.IsAdminResponse, error) {
	isAdmin, err := uh.useCase.IsAdmin(ctx, req.UserName)
	if err != nil {
		return nil, UserErrorsMapper(err)
	}
	return &clientPB.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (uh *UserHandler) ChangeUserPlan(ctx context.Context, req *clientPB.PlanChangeRequest) (*clientPB.PlanChangeResponse, error) {
	plan, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	if !uh.planUserChangeActionMock(ctx, plan) {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	}

	newPlan := roleToUser(req.GetUserRole())

	user, err := uh.useCase.ChangeUserPlan(ctx, plan.UserName, newPlan)
	if err != nil {
		return nil, UserErrorsMapper(err)
	}

	return &clientPB.PlanChangeResponse{
		UserName: user.UserName,
		UserRole: planToUser(user.UserRole),
	}, nil

}
