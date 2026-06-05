package EventActionMockOfOutbox

import (
	"context"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
)

type Event struct {
	AggregateType  string
	AggregateID    string
	EventType      string
	PayLoad        []byte
	IdempotencyKey string
}

type Emmiter interface {
	Emit(ctx context.Context, event Event) error
}

var _ Emmiter = (*MockEmmiter)(nil)

type MockEmmiter struct {
	log LoggerPorts.Logger
}

func NewMockEmmiter(log LoggerPorts.Logger) *MockEmmiter {
	return &MockEmmiter{log: log}
}

func (mem *MockEmmiter) Emit(_ context.Context, event Event) error {
	mem.log.LogInfo("event emmitet (mock)",
		LoggerPorts.Field{Key: "event_type", Value: event.EventType},
		LoggerPorts.Field{Key: "event_id", Value: event.AggregateID},
		LoggerPorts.Field{Key: "aggregate_type", Value: event.AggregateType},
	)
	return nil
}
