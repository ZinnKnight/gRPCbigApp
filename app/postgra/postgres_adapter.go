package postgra

import (
	"context"
	"fmt"
	"gRPCbigapp/app/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		pool: pool,
	}
}

func (or *OrderRepository) DatabaseShemeInitiation(ctx context.Context) error {
	sqlquery := `CREATE TABLE IF NOT EXISTS orders (
    user_Id TEXT NOT NULL,
    order_Id TEXT PRIMARY KEY,
    market_id TEXT NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    order_status TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);`
	_, err := or.pool.Exec(ctx, sqlquery)
	return err
}

func (or *OrderRepository) Create(ctx context.Context, od *domain.OrderDomain) error {
	sqlquery := `INSERT INTO orders (user_Id, order_Id, market_id, price, amount, order_status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := or.pool.Exec(ctx, sqlquery, od.UserID, od.OrderID, od.MarketName, od.Price, od.Amount, od.OrderStatus, od.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка на моменте создания заказа: %w", err)
	}
	return nil
}

func (or *OrderRepository) GetStatus(ctx context.Context, userId string) (string, error) {
	var status string
	err := or.pool.QueryRow(ctx, "SELECT status FROM orders WHERE user_Id = $1", userId).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("получен некорректный статус заказа: %w", err)
	}
	return status, nil
}
