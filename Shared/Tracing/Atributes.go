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

// inside package alice to unbound our export otel/attribute from public

type atributeKeyVal = attribute.KeyValue

func toKeyVal(dataIN []atributeKeyVal) []attribute.KeyValue {
	out := make([]attribute.KeyValue, 0, len(dataIN))

	for _, data := range dataIN {
		out = append(out, data)
	}
	return out
}

func PostgresDB(statment string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", statment),
	}
}

func RedisDB(op string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", op),
	}
}

// For Outbox publisher, if i properly get that - later on need to put that into message broker

func OutboxMesseging(dest string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("messeging.system", "outbox"),
		attribute.String("messeging.destination", dest),
		attribute.String("messeging.operation", "publish"),
	}
}
