package Outbox

import (
	tracing "gRPCbigapp/Shared/Tracing"
	"time"
)

type Event struct {
	ID             int64 // autoincrement
	AggregatorType string
	AggregatorID   string
	EventType      string
	Payload        []byte
	IdempotencyKey string
	CreatedAt      time.Time
	PublishedAt    *time.Time
	RetryCount     int
	TraceContext   tracing.TraceCarrier
}
