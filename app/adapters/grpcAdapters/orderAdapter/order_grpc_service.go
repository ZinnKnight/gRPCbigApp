package orderAdapter

import (
	"context"
	"gRPCbigapp/app/domain"
	orderpb "gRPCbigapp/protofiles/gRPCbigapp/Protofiles/order"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGrpcAdapter struct {
	orderpb.UnimplementedOrderServiceServer
	usecase domain.OrderUsecase
	subs    map[string][]chan *orderpb.OrderUpdate
}

func NewOrderGrpcAdapter(orderUsecase domain.OrderUsecase) *OrderGrpcAdapter {
	return &OrderGrpcAdapter{
		usecase: orderUsecase,
		subs:    make(map[string][]chan *orderpb.OrderUpdate),
	}
}

func (oa *OrderGrpcAdapter) notificationToSubs(orderId, newStatus string) {
	if subs, ok := oa.subs[orderId]; ok {
		updates := &orderpb.OrderUpdate{
			OrderId: orderId,
			Status:  newStatus,
		}
		for _, sub := range subs {
			select {
			case sub <- updates:
			default:
			}
		}
	}
}

func (oa *OrderGrpcAdapter) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	orders, err := oa.usecase.CreateOrder(ctx, domain.CreateOrderParameters{
		UserId:   req.UserId,
		MarketId: req.MarketId,
		UserRole: req.UserRole,
		Price:    req.Price,
		Quantity: req.Quantity,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	oa.notificationToSubs(orders.OrderId, "Создан")

	return &orderpb.CreateOrderResponse{
		OrderId:     orders.OrderId,
		OrderStatus: orders.OrderStatus,
	}, nil
}

func (oa *OrderGrpcAdapter) GetOrderStatus(ctx context.Context, req *orderpb.GetOrderStatusRequest) (*orderpb.GetOrderStatusResponse, error) {
	orderStatus, err := oa.usecase.GetOrderStatus(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}
	return &orderpb.GetOrderStatusResponse{
		OrderStatus: orderStatus,
	}, nil
}

func (oa *OrderGrpcAdapter) StreamOrderUpdates(req *orderpb.StreamOrderRequest, stream orderpb.OrderService_StreamOrderUpdatesServer) error {
	if req.OrderId == "" {
		return status.Error(codes.InvalidArgument, "Необходим OrderId для отображения статуса")
	}
	ch := make(chan *orderpb.OrderUpdate)
	oa.subs[req.OrderId] = append(oa.subs[req.OrderId], ch)

	defer func() {
		subs := oa.subs[req.OrderId]
		for i, sub := range subs {
			if sub == ch {
				subs = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		close(ch)
	}()

	for {
		select {
		case update, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(update); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}
