package Outbox

import (
	"context"
	"fmt"
	"gRPCbigapp/App/Shared/Txmanager"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveEvent(ctx context.Context, event *Event) error {
	const query = `
		INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload, idempotency_key, created_at) 
		VALUES $1, $2, $3, $4, $5, $6`

	var err error
	if tx, ok := Txmanager.ExtractManager(ctx); ok {
		_, err = tx.Exec(ctx, query, event.AggregatorType, event.AggregatorID, event.EventType, event.Payload,
			event.IdempotencyKey, event.CreatedAt,
		)
	} else {
		_, err = tx.Exec(ctx, query, event.AggregatorType, event.AggregatorID, event.EventType, event.Payload,
			event.IdempotencyKey, event.CreatedAt,
		)
	}
	if err != nil {
		return fmt.Errorf("outbox, save event:%w", err)
	}
	return nil
}

func (r *Repository) FetchUnpublished(ctx context.Context, butchSize int) ([]*Event, error) {
	const query = `SELECT id, aggregate_type, aggregate_id, event_type, payload, idempotency_key, created_at,
	retry_count FROM outbox_events WHERE published_at IS NULL ORDER BY id ASC LIMIT $1 FOR UPDATE SKIP LOCKED`

	rows, err := r.pool.Query(ctx, query, butchSize)
	if err != nil {
		return nil, fmt.Errorf("outbox, fetch unpublished events:%w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		e := &Event{}
		if err := rows.Scan(
			&e.ID, &e.AggregatorType, &e.AggregatorID, &e.EventType, &e.Payload, &e.IdempotencyKey, &e.CreatedAt, &e.RetryCount); err != nil {
			return nil, fmt.Errorf("outbox, scan unpublished events:%w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *Repository) MarkPublished(ctx context.Context, eventID int64) error {
	const query = `UPDATE outbox_events SET published_at = retry_count = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, time.Now(), eventID)
	if err != nil {
		return fmt.Errorf("outbox, mark published event:%w", err)
	}
	return nil
}

func (r *Repository) IncrementRetry(ctx context.Context, eventID int64) error {
	const query = `UPDATE outbox_events SET retry_count = retry_count + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, time.Now(), eventID)
	if err != nil {
		return fmt.Errorf("outbox, increment retry:%w", err)
	}
	return nil
}
