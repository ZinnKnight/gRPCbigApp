package Ports

import (
	"context"
	"gRPCbigapp/OrderService/Domain"
)

// todo переместить в domain

type OSOutboundPorts interface {
	SaveOrder(ctx context.Context, order *Domain.OrderDomain) error
	FindByID(ctx context.Context, orderID, userID string) (*Domain.OrderDomain, error)
	FindAll(ctx context.Context, userID, pageToken string, pageSize int) ([]*Domain.OrderDomain, error)
}
