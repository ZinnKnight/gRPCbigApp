package Outbox

import (
	"context"
	"fmt"
	"gRPCbigapp/OrderService/Txmanager"
	"gRPCbigapp/Shared/Events"
	tracing "gRPCbigapp/Shared/Tracing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/codes"
)

type TopicResolver func(e Events.Events) string

var _ Events.Emitter = (*Writer)(nil)

type Writer struct {
	pool  *pgxpool.Pool
	solve TopicResolver
}

func NewWriter(pool *pgxpool.Pool, solve TopicResolver) *Writer {
	return &Writer{
		pool:  pool,
		solve: solve,
	}
}

type dbExecutor interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

func (w *Writer) connection(ctx context.Context) dbExecutor {
	if tx, ok := Txmanager.ExtractManager(ctx); ok {
		return tx
	}
	return w.pool
}

var writeTrace = tracing.Tracer("otbox.writer")

func (w *Writer) Emit(ctx context.Context, e Events.Events) error {
	const query = `
	INSERT INTO outbox (id, aggregate_type, aggregate_id ,event_type ,topic, payload, idempotency_key)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	ctx, span := writeTrace.Start(ctx, "outbox.Emite", tracing.KindProducer)
	defer span.End()

	topic := w.solve(e)
	span.SetAttributes(tracing.OutboxMesseging(topic)...)

	_, err := w.connection(ctx).Exec(
		ctx,
		query,
		uuid.New().String(),
		e.AggregationType,
		e.AggregateId,
		e.EventType,
		topic,
		string(e.PayLoad),
		e.IdempotencyKey,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "outbox.Emit failed")
		return fmt.Errorf("outbox: write event %q: %w", e.EventType, err)
	}
	return nil
}
