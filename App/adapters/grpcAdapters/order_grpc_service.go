package grpcAdapters

import (
	"context"
	"gRPCbigapp/App/domain"
	"gRPCbigapp/App/interceptors"
	orderpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/order"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.OrderDomain) error
	GetStatus(ctx context.Context, orderId string) (string, error)
}

type OrderService struct {
	orderpb.UnimplementedOrderServiceServer
	repa OrderRepository
}

func NewOrderService(repa OrderRepository) *OrderService {
	return &OrderService{repa: repa}
}

func (os *OrderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {

	requestId, ok := ctx.Value(interceptors.RequestID).(string)
	if !ok {
		requestId = ""
	}

	order := domain.OrderDomain{
		UserID:      req.UserId,
		OrderID:     requestId,
		MarketName:  req.MarketId,
		Price:       req.Price,
		Amount:      req.Quantity,
		OrderStatus: "создан",
	}

	err := os.repa.Create(ctx, &order)

	if err != nil {
		return nil, err
	}

	return &orderpb.CreateOrderResponse{
		OrderId:     requestId,
		OrderStatus: "Заказ создан",
	}, nil
}

func (os *OrderService) GetOrderStatus(ctx context.Context, req *orderpb.GetOrderStatusRequest) (*orderpb.GetOrderStatusResponse, error) {
	status, err := os.repa.GetStatus(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &orderpb.GetOrderStatusResponse{OrderStatus: status}, nil
}

// TODO стриминговое отображение статуса заказа
