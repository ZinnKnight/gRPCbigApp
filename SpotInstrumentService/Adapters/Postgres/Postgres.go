package Postgres

import (
	"context"
	"errors"
	"fmt"
	"gRPCbigapp/OrderService/Txmanager"
	tracing "gRPCbigapp/Shared/Tracing"
	"gRPCbigapp/SpotInstrumentService/Domain"
	"gRPCbigapp/SpotInstrumentService/Ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/codes"
)

var trace = tracing.Tracer("db.market_repo")

var _ Ports.SISOutboundRepo = (*MarketRepo)(nil)

type MarketRepo struct {
	pool *pgxpool.Pool
}

func NewSISMarketRepo(pool *pgxpool.Pool) *MarketRepo {
	return &MarketRepo{pool: pool}
}

type dtExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (mr *MarketRepo) connection(ctx context.Context) dtExecutor {
	if tx, ok := Txmanager.ExtractManager(ctx); ok {
		return tx
	}
	return mr.pool
}

func (mr *MarketRepo) FindByName(ctx context.Context, marketId string) (*Domain.MarketDomain, error) {
	const query = `SELECT market_id, market_name, goods_id, accessibility, ttl FROM markets WHERE market_name = $1`

	ctx, span := trace.Start(ctx, "db.FindByID", tracing.KindClient)
	defer span.End()

	span.SetAttributes(tracing.PostgresDB(query)...)

	row := mr.connection(ctx).QueryRow(ctx, query, marketId)

	var m Domain.MarketDomain
	if err := row.Scan(&m.MarketID, &m.MarketName, &m.GoodsID, &m.Accessibility, &m.TTL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			span.AddEvent("market_not_found")
			return nil, Domain.ErrMarketNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "market_repo.FindByID failed")
		return nil, fmt.Errorf("SISMarketRepo, find market FindById: %w", err)
	}
	return &m, nil
}

// для саги, резерв

func (mr *MarketRepo) FindByID(ctx context.Context, marketId string) (*Domain.MarketDomain, error) {
	const query = `
	SELECT market_id, market_name, goods_id, accessibility, ttl FROM markets WHERE market_id = $1
`
	ctx, span := trace.Start(ctx, "db.FindByID", tracing.KindClient)
	defer span.End()

	span.SetAttributes(tracing.PostgresDB(query)...)

	row := mr.connection(ctx).QueryRow(ctx, query, marketId)

	var m Domain.MarketDomain

	if err := row.Scan(&m.MarketID, &m.MarketName, &m.GoodsID, &m.Accessibility, &m.TTL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			span.AddEvent("market_not_found")
			return nil, Domain.ErrMarketNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "market_repo.FindByID failed")
		return nil, fmt.Errorf("SISMarketRepo, find market FindByID: %w", err)
	}
	return &m, nil
}

func (mr *MarketRepo) SaveReserv(ctx context.Context, orderID, marketID, status string) error {
	const query = `
	INSERT INTO reservations (order_id, market_id, status)
	VALUES ($1, $2, $3)
	ON CONFLICT (order_id) DO NOTHING
`
	ctx, span := trace.Start(ctx, "db.SaveReserv", tracing.KindClient)
	defer span.End()

	span.SetAttributes(tracing.PostgresDB(query)...)
	_, err := mr.connection(ctx).Exec(ctx, query, orderID, marketID, status)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "SaveReserv failed")
		return fmt.Errorf("SISMarketRepo, SaveReserv: %w", err)
	}
	return nil
}

func (mr *MarketRepo) FindAll(ctx context.Context, limit int, curs string) ([]*Domain.MarketDomain, error) {
	const query = `SELECT market_id, market_name, goods_id, accessibility, ttl 
	FROM markets 
	WHERE market_id > $1
	ORDER BY market_id ASC
	LIMIT $2`

	ctx, span := trace.Start(ctx, "db.FindAll", tracing.KindClient)
	defer span.End()

	span.SetAttributes(tracing.PostgresDB(query)...)

	rows, err := mr.connection(ctx).Query(ctx, query, curs, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "market_repo.FindAll failed")
		return nil, fmt.Errorf("SISMarketRepo, find markets FindAllMarkets: %w", err)
	}
	defer rows.Close()

	var markets []*Domain.MarketDomain
	for rows.Next() {
		var m Domain.MarketDomain
		if err := rows.Scan(&m.MarketID, &m.MarketName, &m.GoodsID, &m.Accessibility, &m.TTL); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "market_repo.FindAll scan failed")
			return nil, fmt.Errorf("SISMarketRepo, scan markets FindAllMarkets: %w", err)
		}
		markets = append(markets, &m)
	}
	return markets, rows.Err()
}
