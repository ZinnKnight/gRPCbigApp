package CSUseCase

import (
	"context"
	"encoding/json"
	"fmt"
	"gRPCbigapp/App/ClientService/CSDomain"
	"gRPCbigapp/App/ClientService/CSPorts"
	"gRPCbigapp/Shared/EventActionMockOfOutbox"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	tracing "gRPCbigapp/Shared/Tracing"
	"gRPCbigapp/Shared/Txmanager"
	"gRPCbigapp/Shared/ValidationIntercepter"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var csUseCaseTrace = tracing.Tracer("usecase.client_service")

var _ CSPorts.UserInboundPort = (*UserUseCase)(nil)

type UserUseCase struct {
	repo      CSPorts.CSOutboundPorts
	events    EventActionMockOfOutbox.Emmiter
	txManager *Txmanager.TxManager
	logger    LoggerPorts.Logger
}

func NewUserUseCase(repo CSPorts.CSOutboundPorts, event EventActionMockOfOutbox.Emmiter, txManager *Txmanager.TxManager,
	logger LoggerPorts.Logger) *UserUseCase {
	return &UserUseCase{
		events:    event,
		txManager: txManager,
		logger:    logger,
		repo:      repo,
	}
}

func (us *UserUseCase) RegisterUser(ctx context.Context, rui CSPorts.RegisterUserInput) (*CSDomain.User, error) {

	ctx, span := csUseCaseTrace.Start(ctx, "RegisterUser", tracing.KindInternal)
	defer span.End()

	span.SetAttributes(attribute.String("user.name", rui.UserName))

	user := &CSDomain.User{
		UserID:       uuid.New().String(),
		UserName:     rui.UserName,
		UserPassword: rui.UserPassword,
		UserRole:     CSDomain.Free,
	}

	if err := user.ValidateUser(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "client_service.RegisterUser failed")
		return nil, fmt.Errorf("usecase, user registration: %w", err)
	}
	span.SetAttributes(attribute.String("user.id", user.UserID))

	hashedPassword, err := ValidationIntercepter.Hash(user.UserPassword)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "client_service.RegisterUser failed: invalid password hash")
		return nil, fmt.Errorf("usecase, hashing password: %w", err)
	}
	user.UserPassword = hashedPassword

	payload, err := json.Marshal(map[string]interface{}{
		"user_id":   user.UserID,
		"name":      user.UserName,
		"user_plan": string(user.UserRole),
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "json marshal failed")
		return nil, fmt.Errorf("usecase, user marshaling: %w", err)
	}

	event := EventActionMockOfOutbox.Event{
		AggregateType:  "user",
		AggregateID:    user.UserID,
		EventType:      "UserRegistered",
		PayLoad:        payload,
		IdempotencyKey: uuid.New().String(),
	}

	err = us.txManager.Do(ctx, func(ctx context.Context) error {
		if err := us.repo.SaveUser(ctx, user); err != nil {
			return fmt.Errorf("usecase, user saving: %w", err)
		}
		return us.events.Emit(ctx, event)
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "usecase.SaveUser failed")
		us.logger.LogError("Usecase, failed to save user",
			LoggerPorts.Field{Key: "id", Value: user.UserID},
			LoggerPorts.Field{Key: "error", Value: err.Error()},
		)
		return nil, fmt.Errorf("usecase, user registration: %w", err)
	}
	return user, nil
}

func (us *UserUseCase) LoginUser(ctx context.Context, userName, userPassword string) (*CSDomain.User, error) {

	ctx, span := csUseCaseTrace.Start(ctx, "LoginUser", tracing.KindInternal)
	defer span.End()

	span.SetAttributes(attribute.String("user.name", userName))

	user, err := us.repo.GetUser(ctx, userName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "usecase.GetUser failed")
		return nil, fmt.Errorf("usecase, logging user: %w", err)
	}
	if err := ValidationIntercepter.Verify(user.UserPassword, userPassword); err != nil {
		span.SetStatus(codes.Error, "usecase.VerifyUser incorrect credentials")
		return nil, CSDomain.ErrIncorrectCredentials
	}
	return user, nil
}

func (us *UserUseCase) ChangeUserPlan(ctx context.Context, userName string, newPlan CSDomain.UserPlan) (*CSDomain.User, error) {

	ctx, span := csUseCaseTrace.Start(ctx, "ChangeUserPlan", tracing.KindInternal)
	defer span.End()

	span.SetAttributes(attribute.String("user.name", userName), attribute.String("user.plan.new", string(newPlan)))

	var updatedRole *CSDomain.User

	err := us.txManager.Do(ctx, func(ctx context.Context) error {
		user, err := us.repo.GetUser(ctx, userName)
		if err != nil {
			return fmt.Errorf("usecase, look for user: %w", err)
		}
		if err := us.repo.UpdateUserPlan(ctx, user.UserName, newPlan); err != nil {
			return fmt.Errorf("usecase, user plan changing: %w", err)
		}
		user.UserRole = newPlan
		updatedRole = user

		payload, err := json.Marshal(map[string]interface{}{
			"user_name": userName,
			"plan":      string(newPlan),
		})
		if err != nil {
			return fmt.Errorf("usecase, user plan marshaling: %w", err)
		}
		return us.events.Emit(ctx, EventActionMockOfOutbox.Event{
			AggregateType:  "user",
			AggregateID:    userName,
			EventType:      "UserPlanChanged",
			PayLoad:        payload,
			IdempotencyKey: uuid.New().String(),
		})
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "usecase.UpdateUserPlan failed")
		return nil, err
	}

	return updatedRole, nil
}
