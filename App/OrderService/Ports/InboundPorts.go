package Ports

import (
	"context"
	"gRPCbigapp/App/OrderService/Domain"

	"github.com/shopspring/decimal"
)

type OSInboundPort interface {
	CreateOrder(ctx context.Context, cmd CreteOrder) (*Domain.OrderDomain, error)
	GetOrderStatusByID(ctx context.Context, orderID, UserID string) (*Domain.OrderDomain, error)
	GetOrderStatusAll(ctx context.Context, userID, pageToken string, pageSize int) ([]*Domain.OrderDomain, string, error)
}

type CreteOrder struct {
	UserID   string
	MarketID string
	Price    decimal.Decimal
	Quantity decimal.Decimal
}
