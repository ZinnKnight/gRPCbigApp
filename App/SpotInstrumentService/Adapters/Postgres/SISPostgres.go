package Postgres

import (
	"context"
	"fmt"
	"gRPCbigapp/App/Shared/Txmanager"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

func (r *SISMarketRepo) connection(ctx context.Context) dtExecutor {
	if tx, ok := Txmanager.ExtractManager(ctx); ok {
		return tx
	}
	return r.pool
}

func (sisr *SISMarketRepo) FindByID(ctx context.Context, marketId string) (*SISDomain.MarketDomain, error) {
	const query = `SELECT market_id, goods_id, accessibility, ttl FROM markets WHERE market_id = $1`
	row := sisr.connection(ctx).QueryRow(ctx, query, marketId)

	var m SISDomain.MarketDomain
	if err := row.Scan(&m.MarketID, &m.GoodsID, &m.Accessibility, &m.TTL); err != nil {
		if err == pgx.ErrNoRows {
			return nil, SISDomain.ErrMarketNotFound
		}
		return nil, fmt.Errorf("SISMarketRepo, find market FindById: %w", err)
	}
	return &m, nil
}

func (sisr *SISMarketRepo) FindAll(ctx context.Context) ([]*SISDomain.MarketDomain, error) {
	// In future if service will have more colums, probably bad idea, if we not managing them by restrictions
	const query = `SELECT * FROM markets ORDER BY market_id`
	rows, err := sisr.connection(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SISMarketRepo, find markets FindAllMarkets: %w", err)
	}
	defer rows.Close()

	var markets []*SISDomain.MarketDomain
	for rows.Next() {
		var m SISDomain.MarketDomain
		if err := rows.Scan(&m.MarketID, &m.GoodsID, &m.Accessibility, &m.TTL); err != nil {
			return nil, fmt.Errorf("SISMarketRepo, scan markets FindAllMarkets: %w", err)
		}
		markets = append(markets, &m)
	}
	return markets, rows.Err()
}
