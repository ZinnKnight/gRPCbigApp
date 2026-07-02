package Orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"gRPCbigapp/App/OrderService/Domain"
	"gRPCbigapp/Shared/Events"
	"gRPCbigapp/Shared/Idempotentor"
	"gRPCbigapp/Shared/Kafka"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	"gRPCbigapp/Shared/SagaMessages"
	"gRPCbigapp/Shared/Txmanager"

	"github.com/google/uuid"
)

type orderStatusRepository interface {
	UpdateStatus(ctx context.Context, orderID, status string) error
}

type Orchestrator struct {
	repo        orderStatusRepository
	tx          *Txmanager.TxManager
	events      Events.Emitter
	idempotency *Idempotentor.Guard
	logger      LoggerPorts.Logger
}

func NewOrchestrator(repo orderStatusRepository,
	tx *Txmanager.TxManager,
	events Events.Emitter,
	idempotency *Idempotentor.Guard,
	logger LoggerPorts.Logger,
) *Orchestrator {
	return &Orchestrator{
		repo:        repo,
		tx:          tx,
		events:      events,
		idempotency: idempotency,
		logger:      logger,
	}
}

func (o *Orchestrator) Handle(ctx context.Context, event Kafka.Message) error {
	switch event.Header["event_type"] {
	case SagaMessages.EventOrderCreated:
		return o.handleOrderCreated(ctx, event)
	default:
		return nil
	}
}

// подача заказа

func (o *Orchestrator) handleOrderCreated(ctx context.Context, message Kafka.Message) error {
	idempotencyKey := message.Header["idempotency_key"]
	if idempotencyKey == "" {
		return fmt.Errorf("orchestrator,idempotency_key for OrderCreated is empty")
	}

	var created SagaMessages.OrderCreatedPayload

	if err := json.Unmarshal(message.Value, &created); err != nil {
		o.logger.LogError("orchestrator, OrderCreated payload Error",
			LoggerPorts.Field{Key: "error", Value: err.Error()})
		return nil
	}

	return o.tx.Do(ctx, func(ctx context.Context) error {
		first, err := o.idempotency.Acquire(ctx, idempotencyKey)
		if err != nil {
			return err
		}
		if !first {
			o.logger.LogInfo("orchestrator, duplicate OrderCreated skipped",
				LoggerPorts.Field{Key: "order_id", Value: created.OrderID})
			return nil
		}

		cmd := SagaMessages.ReserveStockPayload{
			OrderID:  created.OrderID,
			UserID:   created.UserID,
			MarketID: created.MarketID,
			Amount:   created.Amount,
		}
		payload, err := json.Marshal(cmd)
		if err != nil {
			return fmt.Errorf("orchestrator, marshal payload Error: %w", err)
		}

		return o.events.Emit(ctx, Events.Events{
			AggregationType: "order",
			AggregateId:     created.OrderID,
			EventType:       SagaMessages.CommandReserveStock,
			PayLoad:         payload,
			IdempotencyKey:  uuid.New().String(),
		})
	})
}

// откат
// т.к, по факту откатывать и нечего, прокидываем reject и всё

func (o *Orchestrator) handleReplay(ctx context.Context, msg Kafka.Message, reserved bool) error {
	idempotencyKey := msg.Header["idempotency_key"]
	if idempotencyKey == "" {
		return fmt.Errorf("orchestrator,idempotency_key for Replay is empty")
	}

	orderID, ok := o.orderIDFromReply(msg, reserved)
	if !ok {
		return nil
	}

	newStatus := string(Domain.StatusReserved)
	if reserved {
		newStatus = string(Domain.StatusRejected)
	}

	return o.tx.Do(ctx, func(ctx context.Context) error {
		first, err := o.idempotency.Acquire(ctx, idempotencyKey)
		if err != nil {
			return err
		}
		if !first {
			o.logger.LogInfo("orchestrator, duplicate saga replay skipped",
				LoggerPorts.Field{Key: "order_id", Value: orderID})
			return nil
		}

		if err := o.repo.UpdateStatus(ctx, orderID, newStatus); err != nil {
			return err
		}

		payload, err := json.Marshal(SagaMessages.OrderStatusChangedPayload{
			OrderID: orderID,
			Status:  newStatus,
		})
		if err != nil {
			return fmt.Errorf("orchestrator, marshal payload Error: %w", err)
		}
		o.logger.LogInfo("orchestrator, duplicate saga replay skipped",
			LoggerPorts.Field{Key: "order_id", Value: orderID},
			LoggerPorts.Field{Key: "status", Value: newStatus})

		return o.events.Emit(ctx, Events.Events{
			AggregationType: "order",
			AggregateId:     orderID,
			EventType:       SagaMessages.EventOrderStatusChanged,
			PayLoad:         payload,
			IdempotencyKey:  uuid.New().String(),
		})
	})
}

// order_id из тела

func (o *Orchestrator) orderIDFromReply(message Kafka.Message, reserved bool) (string, bool) {
	if reserved {
		var r SagaMessages.StockPayloadReplay

		if err := json.Unmarshal(message.Value, &r); err != nil {
			o.logger.LogError("orchestrator, orderIDFromReply payload Error",
				LoggerPorts.Field{Key: "error", Value: err.Error()})
			return "", false
		}
		return r.OrderID, true
	}

	var r SagaMessages.StockPayloadReplay
	if err := json.Unmarshal(message.Value, &r); err != nil {
		o.logger.LogError("orchestrator, orderIDFromReply payload Error",
			LoggerPorts.Field{Key: "error", Value: err.Error()})
		return "", false
	}
	return r.OrderID, true
}
