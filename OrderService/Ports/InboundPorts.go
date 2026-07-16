package Ports

import (
	"context"
	"gRPCbigapp/OrderService/Domain"
)

type OSInboundPort interface {
	CreateOrder(ctx context.Context, cmd Domain.CreteOrder) (*Domain.OrderDomain, error)
	GetOrderStatusByID(ctx context.Context, orderID, UserID string) (*Domain.OrderDomain, error)
	GetOrderStatusAll(ctx context.Context, userID, pageToken string, pageSize int) ([]*Domain.OrderDomain, string, error)
}
