package Outbox

import (
	"context"
	"fmt"
	Kafka "gRPCbigapp/Shared/Kafka"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Relay struct {
	pool      *pgxpool.Pool
	producer  *Kafka.Producer
	logger    LoggerPorts.Logger
	batchSize int
	interval  time.Duration
}

func NewRelay(pool *pgxpool.Pool, logger LoggerPorts.Logger, producer *Kafka.Producer, butchsize int, interval time.Duration) *Relay {

	if butchsize <= 0 {
		butchsize = 100
	}

	if interval <= 0 {
		interval = 1 * time.Second
	}

	return &Relay{
		pool:      pool,
		producer:  producer,
		logger:    logger,
		batchSize: butchsize,
		interval:  interval,
	}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.logger.LogInfo("outbox relay starting",
		LoggerPorts.Field{Key: "intervals", Value: r.interval.String()},
		LoggerPorts.Field{Key: "batch_size", Value: r.batchSize},
	)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				n, err := r.BatchRun(ctx)
				if err != nil {
					r.logger.LogError("outbox, relay failed to run",
						LoggerPorts.Field{Key: "error", Value: err.Error()})
					break
				}
				if n < r.batchSize {
					break
				}
			}
		}
	}
}

type outboxRow struct {
	id             string
	topic          string
	key            string
	payLoad        []byte
	eventType      string
	aggType        string
	idempotencyKey string
}

func (r *Relay) BatchRun(ctx context.Context) (int, error) {
	tx, err := r.pool.Begin(ctx)

	if err != nil {
		return 0, fmt.Errorf("outbox relay begin: %w", err)
	}
	defer tx.Rollback(ctx)

	const query = `
	SELECT id,topic, aggregate_id, payload, event_type, aggregate_type, idempotency_key
	FROM outbox
	WHERE published_at IS NULL
	ORDER BY created_at 
	LIMIT $1
	FOR UPDATE SKIP LOCKED`

	rows, err := tx.Query(ctx, query, r.batchSize)
	if err != nil {
		return 0, fmt.Errorf("outbox relay select: %w", err)
	}

	var batch []outboxRow
	for rows.Next() {
		var row outboxRow
		if rows.Scan(&row.id, &row.topic, &row.key, &row.payLoad, &row.eventType, &row.aggType, &row.idempotencyKey); err != nil {
			rows.Close()
			return 0, fmt.Errorf("outbox relay scan: %w", err)
		}
		batch = append(batch, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("outbox, rows iterate: %w", err)
	}
	if len(batch) == 0 {
		return 0, nil
	}

	publishedIDs := make([]string, 0, len(batch))

	for _, row := range batch {
		headers := []Kafka.Header{
			{Key: "event_type", Value: []byte(row.eventType)},
			{Key: "aggregate_type", Value: []byte(row.aggType)},
			{Key: "idempotency_key", Value: []byte(row.idempotencyKey)},
		}
		if err := r.producer.Publish(ctx, row.topic, []byte(row.key), row.payLoad, headers...); err != nil {
			r.logger.LogError("outbox relay publish row",
				LoggerPorts.Field{Key: "topic", Value: row.topic},
				LoggerPorts.Field{Key: "outbox_id", Value: row.id},
				LoggerPorts.Field{Key: "error", Value: err.Error()},
			)
			break
		}
		publishedIDs = append(publishedIDs, row.id)
	}

	if len(publishedIDs) == 0 {
		return 0, nil
	}

	const markSQL = `UPDATE outbox SET published_at = NOW() WHERE id = ANY($1::uuid[])`

	if _, err := tx.Exec(ctx, markSQL, publishedIDs); err != nil {
		return 0, fmt.Errorf("outbox relay mark row: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("outbox, relay commit: %w", err)
	}

	return len(publishedIDs), nil
}
