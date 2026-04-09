package OSPorts

import (
	"context"
	"gRPCbigapp/App/OrderService/OSDomain"
)

type OSInboundPort interface {
	CreteOrder(ctx context.Context, cmd CreteOrder) (string, error)
	GetOrderByID(ctx context.Context, orderID string) (*OSDomain.OrderDomain, error)
	GetAllOrders(ctx context.Context, userID string) ([]*OSDomain.OrderDomain, error)
}

type CreteOrder struct {
	UserID   string
	MarketID string
	Price    float64
	Quantity float64
}
