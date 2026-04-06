package OSUseCase

import (
	"context"
	"gRPCbigapp/OrderService/OSDomain"
	"gRPCbigapp/OrderService/OSPorts/OSOutboundPorts"
)

type OrderService struct {
	repo OSOutboundPorts.OSOutboundPorts
}

func NewOrderService(r OSOutboundPorts.OSOutboundPorts) *OrderService {
	return &OrderService{repo: r}
}

func (os *OrderService) CreateNewOrder(ctx context.Context, order *OSDomain.OrderDomain) error {
	return os.repo.CreateOrder(ctx, order)
}

func (os *OrderService) GetOrderStatusById(ctx context.Context, orderID string) (*OSDomain.OrderDomain, error) {
	return os.repo.OrderStatusByID(ctx, orderID)
}

func (os *OrderService) GetOrderStatusAll(ctx context.Context) (*OSDomain.OrderDomain, error) {
	return os.repo.OrderStatusAll(ctx)
	// probably will change that bcs of JTW
}

func (os *OrderService) StreamOrderStatus(ctx context.Context, orderID string) (chan *OSDomain.OrderDomain, error) {
	return os.repo.StreamOrderStatus(ctx, orderID)
}
