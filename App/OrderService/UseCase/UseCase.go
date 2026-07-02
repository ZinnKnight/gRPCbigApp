package UseCase

import (
	"context"
	"encoding/json"
	"fmt"
	"gRPCbigapp/App/OrderService/Domain"
	"gRPCbigapp/App/OrderService/Ports"
	"gRPCbigapp/Shared/Events"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	"gRPCbigapp/Shared/Policy"
	"gRPCbigapp/Shared/Quota"
	"gRPCbigapp/Shared/Txmanager"

	tracing "gRPCbigapp/Shared/Tracing"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("usecase.order")

var _ Ports.OSInboundPort = (*UseCase)(nil)

type quotaChecker interface {
	Check(ctx context.Context, plan string, action Policy.Action, subject string) (Quota.Decision, error)
}

type UseCase struct {
	repo         Ports.OSOutboundPorts
	events       Events.Emitter
	txManager    *Txmanager.TxManager
	quotaChecker quotaChecker
	logger       LoggerPorts.Logger
}

func NewOSUseCase(repo Ports.OSOutboundPorts,
	event Events.Emitter,
	txManager *Txmanager.TxManager,
	quota quotaChecker,
	logger LoggerPorts.Logger) *UseCase {
	return &UseCase{
		repo:         repo,
		events:       event,
		txManager:    txManager,
		quotaChecker: quota,
		logger:       logger,
	}
}

func (uc *UseCase) enforcedOrderQuota(ctx context.Context, cmd Ports.CreteOrder) error {
	des, err := uc.quotaChecker.Check(ctx, cmd.UserPlan, Policy.ActionCreateOrder, cmd.UserID)
	if err != nil {
		uc.logger.LogError("usecase, order quota check failed (fail-open)",
			LoggerPorts.Field{Key: "user_id", Value: cmd.UserID},
			LoggerPorts.Field{Key: "error", Value: err.Error()},
		)
		return nil
	}
	if !des.Allowed {
		return Domain.ErrOrderQuotaExceeded
	}
	return nil
}

func (uc *UseCase) CreateOrder(ctx context.Context, cmd Ports.CreteOrder) (*Domain.OrderDomain, error) {

	ctx, span := tracer.Start(ctx, "usecase.CreteOrder", tracing.KindInternal)
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", cmd.UserID),
		attribute.String("market.id", cmd.MarketID),
	)

	if err := uc.enforcedOrderQuota(ctx, cmd); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "order quota exceeded")
		return nil, err
	}

	order, err := Domain.NewOrder(cmd.UserID, cmd.MarketID, cmd.Price, cmd.Quantity)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "domain validation")
		return nil, fmt.Errorf("usecase, fail in creating order: %w", err)
	}
	span.SetAttributes(attribute.String("order.id", order.OrderID))

	payload, err := json.Marshal(map[string]interface{}{
		"order_id":  order.OrderID,
		"user_id":   order.UserID,
		"market_id": order.MarketID,
		"price":     order.Price.String(),
		"amount":    order.Amount.String(),
		"status":    string(order.OrderStatus),
	})
	if err != nil {
		return nil, fmt.Errorf("usecase, fail in marshaling order: %w", err)
	}

	event := Events.Events{
		AggregationType: "order",
		AggregateId:     order.OrderID,
		EventType:       "OrderCreated",
		PayLoad:         payload,
		IdempotencyKey:  uuid.New().String(),
	}

	err = uc.txManager.Do(ctx, func(ctx context.Context) error {
		if err := uc.repo.SaveOrder(ctx, order); err != nil {
			return fmt.Errorf("usecase, fail in saving order: %w", err)
		}
		if err := uc.events.Emit(ctx, event); err != nil {
			return fmt.Errorf("usecase, fail in saving event: %w", err)
		}
		return nil
	})
	if err != nil {
		uc.logger.LogError("usecase, fail in creating order: %v, ",
			LoggerPorts.Field{Key: "user_id", Value: order.UserID},
			LoggerPorts.Field{Key: "error", Value: err.Error()},
		)
		return nil, fmt.Errorf("usecase, creating order: %w", err)
	}

	uc.logger.LogInfo("order creted",
		LoggerPorts.Field{Key: "user_id", Value: order.UserID},
		LoggerPorts.Field{Key: "order_id", Value: order.OrderID},
	)

	return order, nil
}

func (uc *UseCase) GetOrderStatusByID(ctx context.Context, orderID, userID string) (*Domain.OrderDomain, error) {
	ctx, span := tracer.Start(ctx, "usecase.GetOrderByID", tracing.KindInternal)
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID), attribute.String("order.id", orderID))

	order, err := uc.repo.FindByID(ctx, orderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repo.FindByID validation")
		return nil, fmt.Errorf("usecase, fail in getting order: %w", err)
	}
	return order, nil
}

func (uc *UseCase) GetOrderStatusAll(ctx context.Context, userID, pageToken string, pageSize int) ([]*Domain.OrderDomain, string, error) {
	ctx, span := tracer.Start(ctx, "usecase.GetAllOrders", tracing.KindInternal)
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID), attribute.Int("page.size", pageSize))

	orders, err := uc.repo.FindAll(ctx, userID, pageToken, pageSize+1)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repo.FindAll validation")
		return nil, "", fmt.Errorf("usecase, fail in getting all orders: %w", err)
	}

	var next string

	if len(orders) > pageSize {
		next = orders[pageSize-1].OrderID
		orders = orders[:pageSize]
	}

	span.SetAttributes(attribute.Int("orders.returned.amount", len(orders)))
	return orders, next, nil
}
