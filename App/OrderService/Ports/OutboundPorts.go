package Ports

import (
	"context"
	"gRPCbigapp/App/OrderService/Domain"
)

type OSOutboundPorts interface {
	SaveOrder(ctx context.Context, order *Domain.OrderDomain) error
	FindByID(ctx context.Context, orderID, userID string) (*Domain.OrderDomain, error)
	FindAll(ctx context.Context, userID, pageToken string, pageSize int) ([]*Domain.OrderDomain, error)
}
