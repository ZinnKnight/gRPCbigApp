package grpcAdapters

import (
	"context"
	"gRPCbigapp/App/adapters/postgra"
	"gRPCbigapp/App/domain"
	orderpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/order"

	"github.com/google/uuid"
)

type OrderService struct {
	orderpb.UnimplementedOrderServiceServer
	repa *postgra.OrderRepoService
}

func NewOrderService(repa *postgra.OrderRepoService) *OrderService {
	return &OrderService{
		repa: repa,
	}
}

func (os *OrderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {

	id := uuid.New().String()

	order := domain.OrderDomain{
		UserID:      req.UserId,
		OrderID:     id,
		MarketName:  req.MarketId,
		Price:       req.Price,
		Amount:      req.Quantity,
		OrderStatus: "создан",
	}

	err := os.repa.Create(ctx, order)

	if err != nil {
		return nil, err
	}

	return &orderpb.CreateOrderResponse{
		OrderId:     id,
		OrderStatus: "Заказ создан",
	}, nil
}

func (os *OrderService) GetOrderStatus(ctx context.Context, req *orderpb.GetOrderStatusRequest) (*orderpb.GetOrderStatusResponse, error) {
	status, err := os.repa.GetStatus(ctx, req.OrderId) // я не могу понять что ему не нравится
	if err != nil {
		return nil, err
	}
	return &orderpb.GetOrderStatusResponse{OrderStatus: status}, nil
}

// TODO стриминговое отображение статуса заказа
