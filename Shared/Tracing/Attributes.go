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

// For Outbox publisher, if i properly get that - later on need to put that into message broker
//Outbox вырезан, но код решил не тереть в 0, что б если что вспомнить чё я делал вообще без залазанья в git
//
//func OutboxMesseging(dest string) []attribute.KeyValue {
//	return []attribute.KeyValue{
//		attribute.String("messaging.system", "outbox"),
//		attribute.String("messaging.destination.name", dest),
//		attribute.String("messaging.operation", "publish"),
//	}
//}
