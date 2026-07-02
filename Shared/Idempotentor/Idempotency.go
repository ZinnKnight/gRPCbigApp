package Idempotentor

import (
	"context"
	"fmt"
	"gRPCbigapp/Shared/Txmanager"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// вынес обработку уникальности в отдельный пакет, т.к возможно он будет меняться

type Guard struct {
	pool     *pgxpool.Pool
	consumer string
}

func NewGuard(pool *pgxpool.Pool, consumer string) *Guard {
	return &Guard{
		pool:     pool,
		consumer: consumer,
	}
}

type dbExecute interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (ga *Guard) connection(ctx context.Context) dbExecute {
	if tx, ok := Txmanager.ExtractManager(ctx); ok {
		return tx
	}
	return ga.pool
}

func (ga *Guard) Acquire(ctx context.Context, idempotencyKey string) (bool, error) {
	const query = `
	INSERT INTO processed_events (consumer, idempotency_key)
	VALUES ($1, $2)
	ON CONFLICT (consumer, idempotency_key) DO NOTHING`

	tag, err := ga.connection(ctx).Exec(ctx, query, ga.consumer, idempotencyKey)
	if err != nil {
		return false, fmt.Errorf("idempotency, acquire transaction: %q, have: %w", idempotencyKey, err)
	}
	return tag.RowsAffected() == 1, nil
}
