package CSUseCase

import (
	"context"
	"encoding/json"
	"fmt"
	"gRPCbigapp/App/ClientService/CSDomain"
	"gRPCbigapp/App/ClientService/CSPorts"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	Outbox2 "gRPCbigapp/Shared/Outbox"
	"gRPCbigapp/Shared/Txmanager"
	"time"

	"github.com/google/uuid"
)

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
	user := &CSDomain.User{
		UserID:       uuid.New().String(),
		UserName:     rui.UserName,
		UserPassword: rui.UserPassword,
		UserRole:     CSDomain.Free,
	}

	if err := user.ValidateUser(); err != nil {
		return nil, fmt.Errorf("usecase, user registration: %w", err)
	}

	payload, err := json.Marshal(map[string]interface{}{
		"user_id":   user.UserID,
		"name":      user.UserName,
		"user_plan": string(user.UserRole),
	})

	if err != nil {
		return nil, fmt.Errorf("usecase, user marshaling: %w", err)
	}

	event := &Outbox2.Event{
		AggregatorType: "user",
		AggregatorID:   user.UserID,
		EventType:      "UserRegistered",
		Payload:        payload,
		IdempotencyKey: uuid.New().String(),
		CreatedAt:      time.Now(),
	}

	err = us.txManager.Do(ctx, func(ctx context.Context) error {
		if err := us.repo.SaveUser(ctx, user); err != nil {
			return fmt.Errorf("usecase, user saving: %w", err)
		}
		return us.outbox.SaveEvent(ctx, event)
	})
	if err != nil {
		us.logger.LogError("Usecase, failed to save user",
			LoggerPorts.Fieled{Key: "id", Value: user.UserID},
			LoggerPorts.Fieled{Key: "error", Value: err.Error()},
		)
		return nil, fmt.Errorf("usecase, user registration: %w", err)
	}
	return user, nil
}

func (us *UserUseCase) LoginUser(ctx context.Context, userID string) (*CSDomain.User, error) {
	user, err := us.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usecase, logging user: %w", err)
	}
	return user, nil
}

func (us *UserUseCase) IsAdmin(ctx context.Context, userID string) (bool, error) {
	return us.repo.IsAdmin(ctx, userID)
}

func (us *UserUseCase) ChangeUserPlan(ctx context.Context, userID string, newPlan CSDomain.UserPlan) error {
	return us.txManager.Do(ctx, func(ctx context.Context) error {
		if err := us.repo.UpdateUserPlan(ctx, userID, newPlan); err != nil {
			return fmt.Errorf("usecase, user plan changing: %w", err)
		}

		payload, err := json.Marshal(map[string]interface{}{
			"user_id": userID,
			"plan":    string(newPlan),
		})
		if err != nil {
			return fmt.Errorf("usecase, user plan marshaling: %w", err)
		}
		return us.outbox.SaveEvent(ctx, &Outbox2.Event{
			AggregatorType: "user",
			AggregatorID:   userID,
			EventType:      "UserPlanChanged",
			Payload:        payload,
			IdempotencyKey: uuid.New().String(),
			CreatedAt:      time.Now(),
		})
	})
}
