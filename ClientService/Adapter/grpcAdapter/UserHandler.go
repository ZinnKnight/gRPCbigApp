package grpcAdapter

import (
	"context"
	"gRPCbigapp/ClientService/Domain"
	"gRPCbigapp/ClientService/Ports"
	"gRPCbigapp/Proto/protoPB"
	"gRPCbigapp/Shared/AuthShared/AuthAdapter"
	"gRPCbigapp/Shared/AuthShared/AuthCTX"
	"gRPCbigapp/Shared/ErrorInterceptor"
	"gRPCbigapp/Shared/Logger/LoggerPorts"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	protoPB.UnimplementedAuthServiceServer
	useCase    Ports.UserInboundPort
	jwt        *AuthAdapter.JWTService
	preRequest PlanUpgradePreRequest
	logger     LoggerPorts.Logger
}

func NewUserhandler(uc Ports.UserInboundPort, log LoggerPorts.Logger, j *AuthAdapter.JWTService, preRequest PlanChangePreRequestStub) *UserHandler {
	return &UserHandler{
		useCase:    uc,
		jwt:        j,
		preRequest: preRequest,
		logger:     log,
	}
}

func planToUser(plan Domain.UserPlan) protoPB.Roles {
	if val, ok := protoPB.Roles_value[string(plan)]; ok {
		return protoPB.Roles(val)
	}
	return protoPB.Roles_UNAUTHORISED_USER
}

func roleToUser(role protoPB.Roles) Domain.UserPlan {
	return Domain.UserPlan(role.String())
}

func (uh *UserHandler) UserRegistration(ctx context.Context, req *protoPB.RegisterRequest) (*protoPB.AuthResponse, error) {
	rui := Ports.RegisterUserInput{
		UserName:     req.GetUserName(),
		UserPassword: req.GetUserPassword(),
	}

	user, err := uh.useCase.UserRegistration(ctx, rui)
	if err != nil {
		uh.logger.LogError("grpc, failed to register user",
			LoggerPorts.Field{Key: "error", Value: err.Error()})
		return nil, err
	}

	token, err := uh.jwt.GenerateToken(user.UserID, user.UserName, string(user.UserRole))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}
	return &protoPB.AuthResponse{Token: token, TokenTtl: uh.jwt.TTLinSeconds()}, nil
}

func (uh *UserHandler) UserLogin(ctx context.Context, req *protoPB.LoginRequest) (*protoPB.AuthResponse, error) {
	user, err := uh.useCase.UserLogin(ctx, req.UserName, req.GetUserPassword())
	if err != nil {
		return nil, err
	}

	token, err := uh.jwt.GenerateToken(user.UserID, user.UserName, string(user.UserRole))
	if err != nil {
		return nil, err
	}
	return &protoPB.AuthResponse{Token: token, TokenTtl: uh.jwt.TTLinSeconds()}, nil
}

func (uh *UserHandler) PlanChange(ctx context.Context, req *protoPB.PlanChangeRequest) (*protoPB.PlanChangeResponse, error) {
	target, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrorInterceptor.NewError(ErrorInterceptor.Unauthenticated, "Требуется авторизация", nil)
	}
	if req.GetUserName() != target.UserName {
		return nil, ErrorInterceptor.NewError(ErrorInterceptor.PermissionDenied, "Возможно изменить лишь собственный тариф", nil)
	}

	newPlan := roleToUser(req.GetUserRole())

	if !Domain.CanSelfPlanChange(newPlan) {
		return nil, ErrorInterceptor.NewError(ErrorInterceptor.PermissionDenied, "Нет возможности изменить на иные тарифы помимо Pro и Free", nil)
	}

	if newPlan == Domain.Pro && !uh.preRequest.UpgradeAgree(ctx, target) {
		return nil, ErrorInterceptor.NewError(ErrorInterceptor.FailedPrecondition, "Для перехода на иной вариант подписки - необходимо выполнение условия", nil)
	}

	user, err := uh.useCase.PlanChange(ctx, target.UserName, newPlan)
	if err != nil {
		return nil, err
	}

	return &protoPB.PlanChangeResponse{
		UserName: user.UserName,
		UserRole: planToUser(user.UserRole),
	}, nil

}
