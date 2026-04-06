package OSOutboundPorts

import (
	"context"
	"gRPCbigapp/App/OrderService/OSDomain"
)

type OSOutboundPorts interface {
	CreateOrder(ctx context.Context, order *OSDomain.OrderDomain) error
	// Create order by add all params ab order

	OrderStatusByID(ctx context.Context, orderID string) (*OSDomain.OrderDomain, error)
	// Grab a specific order by order_id,to check his status

	OrderStatusAll(ctx context.Context) (*OSDomain.OrderDomain, error)
	// Check all orders with thee statuses

	StreamOrderStatus(ctx context.Context, orderID string) (chan *OSDomain.OrderDomain, error)
	// Stream order status of one order
}
