package Postgres

import (
	"context"
	"errors"
	"fmt"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
	tracing "gRPCbigapp/Shared/Tracing"
	"gRPCbigapp/Shared/Txmanager"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/codes"
)

var trace = tracing.Tracer("db.market_repo")

//var _ SISDomain.MarketDomain = (*SISMarketRepo)(nil)

type SISMarketRepo struct {
	pool *pgxpool.Pool
}

func NewSISMarketRepo(pool *pgxpool.Pool) *SISMarketRepo {
	return &SISMarketRepo{pool: pool}
}

type dtExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func (sisr *SISMarketRepo) connection(ctx context.Context) dtExecutor {
	if tx, ok := Txmanager.ExtractManager(ctx); ok {
		return tx
	}
	return sisr.pool
}

func (sisr *SISMarketRepo) FindByID(ctx context.Context, marketId string) (*SISDomain.MarketDomain, error) {
	const query = `SELECT market_id, goods_id, accessibility, ttl FROM markets WHERE market_id = $1`

	ctx, span := trace.Start(ctx, "db.FindByID", tracing.KindClient)
	defer span.End()

	span.SetAttributes(tracing.PostgresDB(query)...)

	row := sisr.connection(ctx).QueryRow(ctx, query, marketId)

	var m SISDomain.MarketDomain
	if err := row.Scan(&m.MarketID, &m.GoodsID, &m.Accessibility, &m.TTL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			span.AddEvent("market_not_found")
			return nil, SISDomain.ErrMarketNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "market_repo.FindByID failed")
		return nil, fmt.Errorf("SISMarketRepo, find market FindById: %w", err)
	}
	return &m, nil
}

func (sisr *SISMarketRepo) FindAll(ctx context.Context, limit int, curs string) ([]*SISDomain.MarketDomain, error) {
	// Change query on more restricted form
	const query = `SELECT market_id, goods_id, accessibility, ttl 
	FROM markets 
	WHERE market_id > $1
	ORDER BY market_id ASC
	LIMIT $2`

	ctx, span := trace.Start(ctx, "db.FindAll", tracing.KindClient)
	defer span.End()

	span.SetAttributes(tracing.PostgresDB(query)...)

	rows, err := sisr.connection(ctx).Query(ctx, query, curs, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "market_repo.FindAll failed")
		return nil, fmt.Errorf("SISMarketRepo, find markets FindAllMarkets: %w", err)
	}
	defer rows.Close()

	var markets []*SISDomain.MarketDomain
	for rows.Next() {
		var m SISDomain.MarketDomain
		if err := rows.Scan(&m.MarketID, &m.GoodsID, &m.Accessibility, &m.TTL); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "market_repo.FindAll scan failed")
			return nil, fmt.Errorf("SISMarketRepo, scan markets FindAllMarkets: %w", err)
		}
		markets = append(markets, &m)
	}
	return markets, rows.Err()
}
