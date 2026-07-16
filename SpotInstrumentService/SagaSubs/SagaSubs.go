package SagaSubs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gRPCbigapp/Shared/Events"
	"gRPCbigapp/Shared/Idempotentor"
	"gRPCbigapp/Shared/Kafka"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	"gRPCbigapp/Shared/SagaMessages"
	"gRPCbigapp/Shared/Txmanager"
	"gRPCbigapp/SpotInstrumentService/Domain"

	"github.com/google/uuid"
)

// частник саги, что производит действия

type merketReserveRepo interface {
	FindByID(ctx context.Context, marketID string) (*Domain.MarketDomain, error)
	SaveReserv(ctx context.Context, orderID, marketID, status string) error
}

type SagaSub struct {
	repo        merketReserveRepo
	tx          *Txmanager.TxManager
	events      Events.Emitter
	idempotency *Idempotentor.Guard
	logger      LoggerPorts.Logger
}

func NewSagaSub(repo merketReserveRepo, tx *Txmanager.TxManager, events Events.Emitter, idempotency *Idempotentor.Guard, logger LoggerPorts.Logger) *SagaSub {
	return &SagaSub{
		repo:        repo,
		tx:          tx,
		events:      events,
		idempotency: idempotency,
		logger:      logger,
	}
}

func (s *SagaSub) HandleReservedStock(ctx context.Context, message Kafka.Message) error {
	if eventT := message.Header["event_type"]; eventT != SagaMessages.CommandReserveStock {
		return nil
	}

	idemKey := message.Header["idempotency_key"]
	if idemKey == "" {
		return fmt.Errorf("SagaSub, no idempotency key")
	}

	var cmd SagaMessages.ReserveStockPayload
	if err := json.Unmarshal(message.Value, &cmd); err != nil {
		s.logger.LogError("sagaSub: bad ReserveStockPayload",
			LoggerPorts.Field{Key: "error", Value: err.Error()})
		return nil
	}

	return s.tx.Do(ctx, func(ctx context.Context) error {
		first, err := s.idempotency.Acquire(ctx, idemKey)
		if err != nil {
			return err
		}
		if !first {
			s.logger.LogInfo("sagaSub: duplicate ReserveStockPayload skipped",
				LoggerPorts.Field{Key: "order_id", Value: cmd.OrderID})
			return nil
		}

		reserved, reason, err := s.reserve(ctx, cmd)
		if err != nil {
			return err
		}

		status := Domain.ReservationReserved

		if !reserved {
			status = Domain.ReservationRejected
		}
		if err := s.repo.SaveReserv(ctx, cmd.OrderID, cmd.MarketID, status); err != nil {
			return err
		}
		return s.emitReply(ctx, cmd, reserved, reason)
	})
}

func (s *SagaSub) reserve(ctx context.Context, cmd SagaMessages.ReserveStockPayload) (bool, string, error) {
	market, err := s.repo.FindByID(ctx, cmd.MarketID)
	if err != nil {
		if errors.Is(err, Domain.ErrMarketNotFound) {
			return false, "market not found", nil
		}
		return false, "", err
	}
	if !market.Accessibility {
		return false, "market not accessible", nil
	}
	return true, "", nil
}

func (s *SagaSub) emitReply(ctx context.Context, cmd SagaMessages.ReserveStockPayload, reserved bool, reason string) error {
	var (
		eventType string
		payLoad   []byte
		err       error
	)

	if reserved {
		eventType = SagaMessages.EventStockReserved
		payLoad, err = json.Marshal(SagaMessages.StockPayloadReplay{
			OrderID:  cmd.OrderID,
			MarketID: cmd.MarketID,
		})
	} else {
		eventType = SagaMessages.EventStockRejected
		payLoad, err = json.Marshal(SagaMessages.StockPayloadReplay{
			OrderID:  cmd.OrderID,
			MarketID: cmd.MarketID,
			Reason:   reason,
		})
	}
	if err != nil {
		return fmt.Errorf("sagaSub: reserve failed: %w", err)
	}

	return s.events.Emit(ctx, Events.Events{
		AggregationType: "order",
		AggregateId:     cmd.OrderID,
		EventType:       eventType,
		PayLoad:         payLoad,
		IdempotencyKey:  uuid.New().String(),
	})
}
