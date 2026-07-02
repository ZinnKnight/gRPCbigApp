package Events

import (
	"context"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
)

// заглушка под некую операцию (например оплату)
// вынес в отдельный пакет, что б отделить от рабочего outbox

type Events struct {
	AggregationType string
	AggregateId     string
	EventType       string
	PayLoad         []byte
	IdempotencyKey  string
}

type Emitter interface {
	Emit(ctx context.Context, event Events) error
}

var _ Emitter = (*MockEmitter)(nil)

type MockEmitter struct {
	log LoggerPorts.Logger
}

func NewMockEmitter(log LoggerPorts.Logger) *MockEmitter {
	return &MockEmitter{
		log: log,
	}
}

func (e *MockEmitter) Emit(_ context.Context, event Events) error {
	e.log.LogInfo("MockEmitter Emit started",
		LoggerPorts.Field{Key: "event_type", Value: event.EventType},
		LoggerPorts.Field{Key: "aggregate_id", Value: event.AggregateId},
		LoggerPorts.Field{Key: "aggregate-type", Value: event.AggregationType},
	)
	return nil
}
