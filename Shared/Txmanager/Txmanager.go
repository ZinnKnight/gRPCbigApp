package Txmanager

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (txm *TxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := txm.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("txManager, started tx: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	defer func() { _ = tx.Rollback(txCtx) }()

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(txCtx); err != nil {
		return fmt.Errorf("txManager, commit tx: %w", err)
	}
	return nil
}

func ExtractManager(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

func (txm *TxManager) Pool() *pgxpool.Pool {
	return txm.pool
}
