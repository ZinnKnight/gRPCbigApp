package Outbox

import "time"

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
}
