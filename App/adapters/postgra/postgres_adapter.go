package postgra

import (
	"context"
	"gRPCbigapp/App/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepoService struct {
	db *pgxpool.Pool
}

func NewOrderRepoService(db *pgxpool.Pool) *OrderRepoService {
	return &OrderRepoService{db: db}
}

func (ors *OrderRepoService) Create(ctx context.Context, order *domain.OrderDomain) error {
	_, err := ors.db.Exec(ctx,
		"INSERT INTO orders (order_id, user_id, market_id, price, quantity, order_status) VALUES ($1, $2, $3, $4, $5, $6) ",
		order.OrderID,
		order.UserID,
		order.MarketName,
		order.Price,
		order.Amount,
		order.OrderStatus,
	)
	return err
}

func (ors *OrderRepoService) GetStatus(ctx context.Context, orderId string) (string, error) {
	var status string

	err := ors.db.QueryRow(ctx,
		"SELECT status FROM orders WHERE order_id = $1",
		orderId,
	).Scan(&status)
	return status, err
}
