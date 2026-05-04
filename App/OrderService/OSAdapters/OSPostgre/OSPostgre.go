package OSPostgre

import (
	"context"
	"errors"
	"fmt"
	"gRPCbigapp/App/OrderService/OSDomain"
	"gRPCbigapp/App/OrderService/OSPorts"
	"gRPCbigapp/Shared/Txmanager"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ OSPorts.OSOutboundPorts = (*OrderRepo)(nil)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{
		pool: pool,
	}
}

type dxExecute interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func (or *OrderRepo) connection(ctx context.Context) dxExecute {
	if tx, ok := Txmanager.ExtractManager(ctx); ok {
		return tx
	}
	return or.pool
}

func (or *OrderRepo) SaveOrder(ctx context.Context, order *OSDomain.OrderDomain) error {
	const query = `INSERT INTO orders(order_id, user_id, market_id, price, amount, order_status, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := or.connection(ctx).Exec(ctx, query, order.OrderID, order.UserID, order.MarketID, order.Price, order.Amount,
		string(order.OrderStatus), order.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres, save order: %w", err)
	}
	return nil
}

func (or *OrderRepo) FindByID(ctx context.Context, orderID, userID string) (*OSDomain.OrderDomain, error) {
	const query = `SELECT order_id, user_id, market_id, price, amount, order_status, created_at 
	FROM orders 
	WHERE order_id = $1 AND user_id = $2`
	rows := or.connection(ctx).QueryRow(ctx, query, orderID, userID)

	var order OSDomain.OrderDomain
	err := rows.Scan(
		&order.OrderID,
		&order.UserID,
		&order.MarketID,
		&order.Price,
		&order.Amount,
		&order.OrderStatus,
		&order.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, OSDomain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("postgres, get order by id: %w", err)
	}
	return &order, nil
}

func (or *OrderRepo) FindAll(ctx context.Context, userID, pageToken string, pageSize int) ([]*OSDomain.OrderDomain, error) {
	const query = `SELECT order_id, user_id, market_id, price, amount, order_status,
       created_at 
	FROM orders 
	WHERE user_id = $1 AND order_id > $2
	ORDER BY order_id ASC 
	LIMIT $3`
	rows, err := or.connection(ctx).Query(ctx, query, userID, pageToken, pageSize)
	if err != nil {
		return nil, fmt.Errorf("postgres, get all orders: %w", err)
	}
	defer rows.Close()

	var orders []*OSDomain.OrderDomain
	for rows.Next() {
		var ord OSDomain.OrderDomain

		if err := rows.Scan(&ord.OrderID, &ord.UserID, &ord.MarketID, &ord.Price, &ord.Amount, &ord.OrderStatus, &ord.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres, scan for all orders: %w", err)
		}
		orders = append(orders, &ord)
	}
	return orders, rows.Err()
}
