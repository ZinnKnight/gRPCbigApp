package postgra

import (
	"context"

	"gRPCbigapp/App/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepoService struct {
	databasa *pgxpool.Pool
}

func NewOrderRepoService(databasa *pgxpool.Pool) *OrderRepoService {
	return &OrderRepoService{databasa: databasa}
}

func (r *OrderRepoService) Create(ctx context.Context, order *domain.OrderDomain) error {
	_, err := r.databasa.Exec(ctx,
		"INSERT INTO orders(id,user_id,market_id,price,quantity,status) VALUES($1,$2,$3,$4,$5,$6)",
		order.OrderID, order.UserID, order.MarketName, order.Price, order.Amount, order.OrderStatus,
	)
	return err
}

func (r *OrderRepoService) GetStatus(ctx context.Context, id int64) (string, error) {
	var status string
	err := r.databasa.QueryRow(ctx, "SELECT status FROM orders WHERE id=$1", id).Scan(&status)
	return status, err
}
