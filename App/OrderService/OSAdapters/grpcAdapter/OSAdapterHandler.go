package grpcAdapter

import (
	"context"
	"gRPCbigapp/App/OrderService/OSPorts"
	orderpb "gRPCbigapp/Proto/order"
	"gRPCbigapp/Shared/Auth/AuthCTX"
	"gRPCbigapp/Shared/Logger/LoggerPorts"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect price")
	}

	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect quantity")
	}

	cmd := OSPorts.CreteOrder{
		UserID:   user.UserID,
		MarketID: req.MarketId,
		Price:    price,
		Quantity: quantity,
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
	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}
	order, err := o.useCase.GetOrderByID(ctx, req.OrderId, user.UserID)
	if err != nil {
		return nil, DomainErrorMapping(err)
	}
	return &orderpb.OrderStatusByIDResponse{
		OrderId:     order.OrderID,
		OrderStatus: string(order.OrderStatus),
	}, nil
}

func (o *OrderHandler) GetAllOrderStatuses(ctx context.Context, req *orderpb.OrderStatusAllRequest) (*orderpb.OrderStatusAllResponse, error) {
	size := int(req.PageSize)
	if size <= 0 || size > 100 {
		size = 20
	}

	curs := req.PageToken

	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}

	orders, nextPageToken, err := o.useCase.GetAllOrders(ctx, user.UserID, curs, size)
	if err != nil {
		return nil, DomainErrorMapping(err)
	}

	protoOrders := make([]*orderpb.OrderStatusStruct, 0, len(orders))

	for _, ord := range orders {
		protoOrders = append(protoOrders, &orderpb.OrderStatusStruct{
			OrderId: ord.OrderID,
			Status:  string(ord.OrderStatus),
		})
	}

	return &orderpb.OrderStatusAllResponse{
		Orders:        protoOrders,
		NextPageToken: nextPageToken,
	}, nil
}
