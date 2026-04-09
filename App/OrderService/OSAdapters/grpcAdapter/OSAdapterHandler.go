package grpcAdapter

import (
	"context"
	"gRPCbigapp/App/OrderService/OSPorts"
	"gRPCbigapp/App/Shared/Auth/AuthCTX"
	"gRPCbigapp/App/Shared/Logger/LoggerPorts"
	orderpb "gRPCbigapp/Proto/order"
)

type OrderHandler struct {
	orderpb.UnimplementedOrderServiceServer
	useCase OSPorts.OSInboundPort
	logger  LoggerPorts.Logger
}

func NewOrderHandler(log LoggerPorts.Logger, osp OSPorts.OSInboundPort) *OrderHandler {
	return &OrderHandler{
		logger:  log,
		useCase: osp,
	}
}

func (o *OrderHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}

	cmd := OSPorts.CreteOrder{
		UserID:   user.UserID,
		MarketID: req.MarketId,
		Price:    req.Price,
		Quantity: req.Quantity,
	}

	orderID, err := o.useCase.CreteOrder(ctx, cmd)
	if err != nil {
		o.logger.LogError("grpc, failed to crete order",
			LoggerPorts.Fieled{Key: "id", Value: user.UserID},
			LoggerPorts.Fieled{Key: "error", Value: err.Error()})
		return nil, DomainErrorMapping(err)
	}

	return &orderpb.CreateOrderResponse{OrderId: orderID, OrderStatus: "Created"}, nil
}

func (o *OrderHandler) GetOrderStatusByID(ctx context.Context, req *orderpb.OrderStatusByIDRequest) (*orderpb.OrderStatusByIDResponse, error) {
	order, err := o.useCase.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		return nil, DomainErrorMapping(err)
	}
	return &orderpb.OrderStatusByIDResponse{
		OrderId:     order.OrderID,
		OrderStatus: string(order.OrderStatus),
	}, nil
}

func (o *OrderHandler) GetAllOrderStatuses(ctx context.Context, req *orderpb.OrderStatusAllRequest) (*orderpb.OrderStatusAllResponse, error) {
	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}
	orders, err := o.useCase.GetAllOrders(ctx, user.UserID)
	if err != nil {
		return nil, DomainErrorMapping(err)
	}

	var ids, statuses []string
	for _, ord := range orders {
		ids = append(ids, ord.OrderID)
		statuses = append(statuses, string(ord.OrderStatus))
	}
	return &orderpb.OrderStatusAllResponse{
		OrderId:     ids,
		OrderStatus: statuses,
	}, nil
}
