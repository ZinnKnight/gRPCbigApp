package CSUseCase

import (
	"context"
	"encoding/json"
	"fmt"
	"gRPCbigapp/App/ClientService/CSDomain"
	"gRPCbigapp/App/ClientService/CSPorts"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	Outbox2 "gRPCbigapp/Shared/Outbox"
	tracing "gRPCbigapp/Shared/Tracing"
	"gRPCbigapp/Shared/Txmanager"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var csUseCaseTrace = tracing.Tracer("usecase.client_service")

var _ CSPorts.UserInboundPort = (*UserUseCase)(nil)

type UserUseCase struct {
	repo      CSPorts.CSOutboundPorts
	outbox    *Outbox2.Repository
	txManager *Txmanager.TxManager
	logger    LoggerPorts.Logger
}

func NewUserUseCase(repo CSPorts.CSOutboundPorts, outbox *Outbox2.Repository, txManager *Txmanager.TxManager,
	logger LoggerPorts.Logger) *UserUseCase {
	return &UserUseCase{
		outbox:    outbox,
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

	event := &Outbox2.Event{
		AggregatorType: "user",
		AggregatorID:   user.UserID,
		EventType:      "UserRegistered",
		Payload:        payload,
		IdempotencyKey: uuid.New().String(),
		CreatedAt:      time.Now(),
		TraceContext:   tracing.PlaceIntoCar(ctx),
	}

	err = us.txManager.Do(ctx, func(ctx context.Context) error {
		if err := us.repo.SaveUser(ctx, user); err != nil {
			return fmt.Errorf("usecase, user saving: %w", err)
		}
		return us.outbox.SaveEvent(ctx, event)
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

func (us *UserUseCase) LoginUser(ctx context.Context, userID string) (*CSDomain.User, error) {

	ctx, span := csUseCaseTrace.Start(ctx, "LoginUser", tracing.KindInternal)
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID))

	user, err := us.repo.GetUser(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "usecase.GetUser failed")
		return nil, fmt.Errorf("usecase, logging user: %w", err)
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
			span.RecordError(err)
			span.SetStatus(codes.Error, "usecase.GetUser failed")
			return fmt.Errorf("usecase, changing user plan: %w", err)
		}
		if err := us.repo.UpdateUserPlan(ctx, user.UserName, newPlan); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "usecase.UpdateUserPlan failed")
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
		return us.outbox.SaveEvent(ctx, &Outbox2.Event{
			AggregatorType: "user",
			AggregatorID:   userName,
			EventType:      "UserPlanChanged",
			Payload:        payload,
			IdempotencyKey: uuid.New().String(),
			CreatedAt:      time.Now(),
			TraceContext:   tracing.PlaceIntoCar(ctx),
		})
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "usecase.UpdateUserPlan failed")
		return nil, fmt.Errorf("usecase, user plan changing: %w", err)
	}

	return updatedRole, nil
}
