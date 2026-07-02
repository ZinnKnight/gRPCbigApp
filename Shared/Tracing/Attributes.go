package Tracing

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Wrapers for not grab whole pakage

var (
	KindServer   = trace.WithSpanKind(trace.SpanKindServer)
	KindClient   = trace.WithSpanKind(trace.SpanKindClient)
	KindInternal = trace.WithSpanKind(trace.SpanKindInternal)
	KindProducer = trace.WithSpanKind(trace.SpanKindProducer)
	KindConsumer = trace.WithSpanKind(trace.SpanKindConsumer)
)

func PostgresDB(statement string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", statement),
	}
}

func RedisDB(op string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", op),
	}
}

func OutboxMesseging(dest string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("messaging.system", "outbox"),
		attribute.String("messaging.destination.name", dest),
		attribute.String("messaging.operation", "publish"),
	}
}
