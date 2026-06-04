package grpcAdapter

import (
	"context"
	"gRPCbigapp/App/ClientService/CSDomain"
	"gRPCbigapp/App/ClientService/CSPorts"
	"gRPCbigapp/Proto/protoPB/clientPB"
	"gRPCbigapp/Shared/Auth/AuthAdapter"
	"gRPCbigapp/Shared/Auth/AuthCTX"
	"gRPCbigapp/Shared/ErrorInterceptor"
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

func planToUser(plan CSDomain.UserPlan) clientPB.Roles {
	if val, ok := clientPB.Roles_value[string(plan)]; ok {
		return clientPB.Roles(val)
	}
	return clientPB.Roles_UNAUTHORISED_USER
}

func roleToUser(role clientPB.Roles) CSDomain.UserPlan {
	return CSDomain.UserPlan(role.String())
}

// Это временный мок, как будет кафка - необходимо будет переделать
func (uh *UserHandler) planChangeActionCompleted(_ context.Context, _ *AuthCTX.UserAuth) bool {
	return true
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
		return nil, err
	}

	token, err := uh.jwt.GenerateToken(user.UserID, user.UserName, string(user.UserRole))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}
	return &clientPB.AuthResponse{Token: token, TokenTtl: uh.jwt.TTLinSeconds()}, nil
}

func (uh *UserHandler) UserLogin(ctx context.Context, req *clientPB.LoginRequest) (*clientPB.AuthResponse, error) {
	user, err := uh.useCase.LoginUser(ctx, req.UserName, req.GetUserPassword())
	if err != nil {
		return nil, err
	}

	token, err := uh.jwt.GenerateToken(user.UserID, user.UserName, string(user.UserRole))
	if err != nil {
		return nil, err
	}
	return &clientPB.AuthResponse{Token: token, TokenTtl: uh.jwt.TTLinSeconds()}, nil
}

func (uh *UserHandler) ChangeUserPlan(ctx context.Context, req *clientPB.PlanChangeRequest) (*clientPB.PlanChangeResponse, error) {
	target, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrorInterceptor.NewError(ErrorInterceptor.Unauthenticated, "Требуется авторизация", nil)
	}
	if !uh.planChangeActionCompleted(ctx, target) {
		return nil, ErrorInterceptor.NewError(ErrorInterceptor.FailedPrecondition, "Не выполенны необходимые шаги", nil)
	}

	newPlan := roleToUser(req.GetUserRole())

	user, err := uh.useCase.ChangeUserPlan(ctx, target.UserName, newPlan)
	if err != nil {
		return nil, err
	}

	return &clientPB.PlanChangeResponse{
		UserName: user.UserName,
		UserRole: planToUser(user.UserRole),
	}, nil

}
