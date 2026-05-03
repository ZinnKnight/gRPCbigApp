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

	"github.com/google/uuid"
)

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
	order, err := OSDomain.NewOrder(cmd.UserID, cmd.MarketID, cmd.Price, cmd.Quantity)
	if err != nil {
		return "", fmt.Errorf("usecase, fail in creating order: %w", err)
	}

	payload, err := json.Marshal(map[string]interface{}{
		"order_id":  order.OrderID,
		"user_id":   order.UserID,
		"market_id": order.MarketID,
		"price":     order.Price.String(),
		"amount":    order.Amount.String(),
		"status":    string(order.OrderStatus),
	})
	if err != nil {
		return "", fmt.Errorf("usecase, fail in marshaling order: %v", err)
	}

	event := &Outbox2.Event{
		AggregatorType: "order",
		AggregatorID:   order.OrderID,
		EventType:      "OrderCreated",
		Payload:        payload,
		IdempotencyKey: uuid.New().String(),
		CreatedAt:      time.Now(),
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
			LoggerPorts.Fieled{Key: "user_id", Value: order.UserID},
			LoggerPorts.Fieled{Key: "error", Value: err.Error()},
		)
		return "", fmt.Errorf("usecase, creating order: %w", err)
	}

	osu.logger.LogInfo("order creted",
		LoggerPorts.Fieled{Key: "user_id", Value: order.UserID},
		LoggerPorts.Fieled{Key: "order_id", Value: order.OrderID},
	)

	return order.OrderID, nil
}

func (osu *OSUseCase) GetOrderByID(ctx context.Context, orderID, userID string) (*OSDomain.OrderDomain, error) {
	order, err := osu.repo.FindByID(ctx, orderID, userID)
	if err != nil {
		return nil, fmt.Errorf("usecase, fail in getting order: %v", err)
	}
	return order, nil
}

func (osu *OSUseCase) GetAllOrders(ctx context.Context, userID, pageToken string, pageSize int) ([]*OSDomain.OrderDomain, string, error) {
	orders, err := osu.repo.FindAll(ctx, userID, pageToken, pageSize+1)
	if err != nil {
		return nil, "", fmt.Errorf("usecase, fail in getting all orders: %v", err)
	}

	var next string

	if len(orders) > pageSize {
		next = orders[pageSize-1].OrderID
		orders = orders[:pageSize]
	}

	return orders, next, nil
}
