package OSPorts

import (
	"context"
	"gRPCbigapp/App/OrderService/OSDomain"

	"github.com/shopspring/decimal"
)

type OSInboundPort interface {
	CreteOrder(ctx context.Context, cmd CreteOrder) (string, error)
	GetOrderByID(ctx context.Context, orderID, UserID string) (*OSDomain.OrderDomain, error)
	GetAllOrders(ctx context.Context, userID, pageToken string, pageSize int) ([]*OSDomain.OrderDomain, string, error)
}

type CreteOrder struct {
	UserID   string
	MarketID string
	Price    decimal.Decimal
	Quantity decimal.Decimal
}
