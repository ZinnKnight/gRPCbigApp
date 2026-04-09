package OSPorts

import (
	"context"
	"gRPCbigapp/App/OrderService/OSDomain"
)

type OSOutboundPorts interface {
	SaveOrder(ctx context.Context, order *OSDomain.OrderDomain) error
	FindByID(ctx context.Context, orderID string) (*OSDomain.OrderDomain, error)
	FindAll(ctx context.Context, userID string) ([]*OSDomain.OrderDomain, error)
}
