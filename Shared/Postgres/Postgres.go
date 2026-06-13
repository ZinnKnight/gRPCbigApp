package Postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// эта часть как раз принимает то что мы насоздавали в Config и на основе его собираем тут что б разгрузить main

type Config struct {
	DatabaseURL       string
	MaxConnections    int32
	MinConnections    int32
	MaxConnTTL        time.Duration
	DBMaxConnIdTTL    time.Duration
	HealthCheckPeriod time.Duration
	AfterConn         func(ctx context.Context, conn *pgx.Conn) error
}

func NewPool(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database url: %w", err)
	}

	if config.MaxConnections > 0 {
		poolConfig.MaxConns = config.MaxConnections
	}
	if config.MinConnections > 0 {
		poolConfig.MinConns = config.MinConnections
	}
	if config.MaxConnTTL > 0 {
		poolConfig.MaxConnLifetime = config.MaxConnTTL
	}
	if config.DBMaxConnIdTTL > 0 {
		poolConfig.MaxConnIdleTime = config.DBMaxConnIdTTL
	}
	if config.AfterConn != nil {
		poolConfig.AfterConnect = config.AfterConn
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return pool, nil
}
