package Outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"gRPCbigapp/Shared/Txmanager"
	"time"

	tracing "gRPCbigapp/Shared/Tracing"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/codes"
)

var tracer = tracing.Tracer("outbox.repository")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveEvent(ctx context.Context, event *Event) error {
	ctx, span := tracer.Start(ctx, "outbox.SaveEvent", tracing.KindClient)
	defer span.End()

	const query = `
		INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload, idempotency_key, created_at, trace_context) 
		VALUES $1, $2, $3, $4, $5, $6, $7`

	span.SetAttributes(tracing.PostgresDB(query)...)

	ctxTraceJSON, err := json.Marshal(event.TraceContext)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "marshal ctx json fail")
		return fmt.Errorf("outbox, marshal error in context: %w", err)
	}

	if tx, ok := Txmanager.ExtractManager(ctx); ok {
		_, err = tx.Exec(ctx, query, event.AggregatorType, event.AggregatorID, event.EventType, event.Payload,
			event.IdempotencyKey, event.CreatedAt, ctxTraceJSON,
		)
	} else {
		_, err = tx.Exec(ctx, query, event.AggregatorType, event.AggregatorID, event.EventType, event.Payload,
			event.IdempotencyKey, event.CreatedAt, ctxTraceJSON,
		)
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "execution event fail")
		return fmt.Errorf("outbox, save event:%w", err)
	}
	return nil
}

func (r *Repository) FetchUnpublished(ctx context.Context, butchSize int) ([]*Event, error) {
	ctx, span := tracer.Start(ctx, "outbox.FetchUnpublished", tracing.KindClient)
	defer span.End()

	const query = `SELECT id, aggregate_type, aggregate_id, event_type, payload, idempotency_key, created_at,
	retry_count, trace_context FROM outbox_events WHERE published_at IS NULL ORDER BY id ASC LIMIT $1 FOR UPDATE SKIP LOCKED`

	span.SetAttributes(tracing.PostgresDB(query)...)

	rows, err := r.pool.Query(ctx, query, butchSize)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "query fetch unpublished fail")
		return nil, fmt.Errorf("outbox, fetch unpublished events:%w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		e := &Event{}
		var ctxTraceJSON []byte
		if err := rows.Scan(
			&e.ID, &e.AggregatorType, &e.AggregatorID,
			&e.EventType, &e.Payload, &e.IdempotencyKey,
			&e.CreatedAt, &e.RetryCount, &ctxTraceJSON); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "rows scan fail")
			return nil, fmt.Errorf("outbox, scan unpublished events:%w", err)
		}

		if len(ctxTraceJSON) > 0 {
			car := tracing.TraceCarier{}
			if jerr := json.Unmarshal(ctxTraceJSON, &car); jerr == nil {
				e.TraceContext = car
			}
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *Repository) MarkPublished(ctx context.Context, eventID int64) error {
	const query = `UPDATE outbox_events SET published_at = retry_count = $1 WHERE id = $2`

	ctx, span := tracer.Start(ctx, "outbox.MarkPublished", tracing.KindClient)
	defer span.End()
	span.SetAttributes(tracing.PostgresDB(query)...)

	_, err := r.pool.Exec(ctx, query, time.Now(), eventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "mark publisher execution fail")
		return fmt.Errorf("outbox, mark published event:%w", err)
	}
	return nil
}

func (r *Repository) IncrementRetry(ctx context.Context, eventID int64) error {
	const query = `UPDATE outbox_events SET retry_count = retry_count + 1 WHERE id = $1`

	ctx, span := tracer.Start(ctx, "outbox.IncrementRetry", tracing.KindClient)
	defer span.End()
	span.SetAttributes(tracing.PostgresDB(query)...)

	_, err := r.pool.Exec(ctx, query, time.Now(), eventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "outbox, increment retry fail")
		return fmt.Errorf("outbox, increment retry:%w", err)
	}
	return nil
}
