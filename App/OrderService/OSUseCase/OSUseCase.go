package OSUseCase

import (
	"context"
	"encoding/json"
	"fmt"
	"gRPCbigapp/App/OrderService/OSDomain"
	"gRPCbigapp/App/OrderService/OSPorts"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	Outbox2 "gRPCbigapp/Shared/Outbox"
	"gRPCbigapp/Shared/Txmanager"
	"time"

	tracing "gRPCbigapp/Shared/Tracing"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// weird name, but done to save a logick in Jaeger for filter
var tracer = otel.Tracer("usecase.order")

var _ OSPorts.OSInboundPort = (*OSUseCase)(nil)

type OSUseCase struct {
	repo      OSPorts.OSOutboundPorts
	outbox    *Outbox2.Repository
	txManager *Txmanager.TxManager
	logger    LoggerPorts.Logger
}

func NewOSUseCase(repo OSPorts.OSOutboundPorts, outbox *Outbox2.Repository, txManager *Txmanager.TxManager, logger LoggerPorts.Logger) *OSUseCase {
	return &OSUseCase{
		repo:      repo,
		outbox:    outbox,
		txManager: txManager,
		logger:    logger,
	}
}

func (osu *OSUseCase) CreteOrder(ctx context.Context, cmd OSPorts.CreteOrder) (string, error) {

	ctx, span := tracer.Start(ctx, "usecase.CreteOrder", tracing.KindInternal)
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", cmd.UserID),
		attribute.String("market.id", cmd.MarketID),
	)

	order, err := OSDomain.NewOrder(cmd.UserID, cmd.MarketID, cmd.Price, cmd.Quantity)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "domain validation")
		return "", fmt.Errorf("usecase, fail in creating order: %w", err)
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
		return "", fmt.Errorf("usecase, fail in marshaling order: %w", err)
	}

	event := &Outbox2.Event{
		AggregatorType: "order",
		AggregatorID:   order.OrderID,
		EventType:      "OrderCreated",
		Payload:        payload,
		IdempotencyKey: uuid.New().String(),
		CreatedAt:      time.Now(),
		TraceContext:   tracing.PlaceIntoCar(ctx),
	}

	err = osu.txManager.Do(ctx, func(ctx context.Context) error {
		if err := osu.repo.SaveOrder(ctx, order); err != nil {
			return fmt.Errorf("usecase, fail in saving order: %w", err)
		}
		if err := osu.outbox.SaveEvent(ctx, event); err != nil {
			return fmt.Errorf("usecase, fail in saving event: %w", err)
		}
		return nil
	})
	if err != nil {
		osu.logger.LogError("usecase, fail in creating order: %v, ",
			LoggerPorts.Field{Key: "user_id", Value: order.UserID},
			LoggerPorts.Field{Key: "error", Value: err.Error()},
		)
		return "", fmt.Errorf("usecase, creating order: %w", err)
	}

	osu.logger.LogInfo("order creted",
		LoggerPorts.Field{Key: "user_id", Value: order.UserID},
		LoggerPorts.Field{Key: "order_id", Value: order.OrderID},
	)

	return order.OrderID, nil
}

func (osu *OSUseCase) GetOrderByID(ctx context.Context, orderID, userID string) (*OSDomain.OrderDomain, error) {
	ctx, span := tracer.Start(ctx, "usecase.GetOrderByID", tracing.KindInternal)
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID), attribute.String("order.id", orderID))

	order, err := osu.repo.FindByID(ctx, orderID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repo.FindByID validation")
		return nil, fmt.Errorf("usecase, fail in getting order: %w", err)
	}
	return order, nil
}

func (osu *OSUseCase) GetAllOrders(ctx context.Context, userID, pageToken string, pageSize int) ([]*OSDomain.OrderDomain, string, error) {
	ctx, span := tracer.Start(ctx, "usecase.GetAllOrders", tracing.KindInternal)
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID), attribute.Int("page.size", pageSize))

	orders, err := osu.repo.FindAll(ctx, userID, pageToken, pageSize+1)
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
