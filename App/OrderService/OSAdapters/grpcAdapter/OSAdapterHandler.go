package grpcAdapter

import (
	"context"
	"gRPCbigapp/App/OrderService/OSDomain"
	"gRPCbigapp/App/OrderService/OSPorts"
	orderpb "gRPCbigapp/Proto/protoPB/orderPB"
	"gRPCbigapp/Shared/Auth/AuthCTX"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	"time"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	moneyconverter "gRPCbigapp/Shared/Converters/Money"
)

// т.к outbox вырезал, а ломать всю логику не хочется - поставил вот такую затычку

const poolInterval = 5 * time.Second

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

// сразу просматриваем статусы/конвертеры из pb.proto

func pbProtoStatuses(status OSDomain.OrderStatus) orderpb.OrderStatus {
	if val, ok := orderpb.OrderStatus_value[string(status)]; ok {
		return orderpb.OrderStatus(val)
	}
	return orderpb.OrderStatus_UNREGISTERED_STATUS
}

func pbOrderConverter(convert *OSDomain.OrderDomain) *orderpb.Order {
	return &orderpb.Order{
		UserId:      convert.UserID,
		OrderId:     convert.OrderID,
		MarketId:    convert.MarketID,
		Price:       moneyconverter.DecToMoney(convert.Price, moneyconverter.Currency),
		Amount:      moneyconverter.DecimalToDecimalPB(convert.Amount),
		OrderStatus: pbProtoStatuses(convert.OrderStatus),
		CreatedAt:   convert.CreatedAt.Unix(),
	}
}

func (o *OrderHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}

	price := moneyconverter.MoneyToDec(req.GetPrice())

	amount, err := moneyconverter.DecimalPBToDecimal(req.GetAmount())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "incorrect amount")
	}

	cmd := OSPorts.CreteOrder{
		UserID:   user.UserID,
		MarketID: req.GetMarketId(),
		Price:    price,
		Quantity: amount,
	}

	orderID, err := o.useCase.CreteOrder(ctx, cmd)
	if err != nil {
		o.logger.LogError("grpc, failed to crete order",
			LoggerPorts.Field{Key: "id", Value: user.UserID},
			LoggerPorts.Field{Key: "error", Value: err.Error()})
		return nil, DomainErrorMapping(err)
	}

	return &orderpb.CreateOrderResponse{CreateOrderResponse: pbOrderConverter(orderID)}, nil
}

func (o *OrderHandler) GetOrderStatusByID(ctx context.Context, req *orderpb.OrderStatusByIDRequest) (*orderpb.OrderStatusByIDResponse, error) {
	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}
	order, err := o.useCase.GetOrderByID(ctx, req.GetOrderId(), user.UserID)
	if err != nil {
		return nil, DomainErrorMapping(err)
	}
	return &orderpb.OrderStatusByIDResponse{
		OrderStatusResponse: pbOrderConverter(order),
	}, nil
}

func (o *OrderHandler) GetAllOrderStatuses(ctx context.Context, req *orderpb.OrderStatusAllRequest) (*orderpb.OrderStatusAllResponse, error) {
	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}

	size := int(req.GetPageSize())
	if size <= 0 || size > 100 {
		size = 20
	}

	orders, nextPageToken, err := o.useCase.GetAllOrders(ctx, user.UserID, req.GetPageToken(), size)
	if err != nil {
		return nil, DomainErrorMapping(err)
	}

	protoOrders := make([]*orderpb.Order, 0, len(orders))

	for _, ord := range orders {
		protoOrders = append(protoOrders, pbOrderConverter(ord))
	}

	return &orderpb.OrderStatusAllResponse{
		AllOrdersStatusesResponse: protoOrders,
		NextPageToken:             nextPageToken,
	}, nil
}

func (o *OrderHandler) StreamUpdateOrder(req *orderpb.StreamOrderRequest, stream orderpb.OrderService_StreamOrderUpdatesServer) error {
	ctx := stream.Context()

	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return ErrUnauthenticated
	}

	orderID := req.GetOrderId()

	var lastSend OSDomain.OrderStatus
	firstSend := true

	send := func() (terminal bool, err error) {
		order, err := o.useCase.GetOrderByID(ctx, orderID, user.UserID)

		if err != nil {
			return false, DomainErrorMapping(err)
		}

		if firstSend || order.OrderStatus != lastSend {
			if err := stream.Send(&orderpb.OrderStatusByIDResponse{
				OrderStatusResponse: pbOrderConverter(order)}); err != nil {
				// пока оставил так, позже буду мапить ошибки через отдельный обработчик
				return false, err
			}
			lastSend = order.OrderStatus
			firstSend = false
		}
		return order.OrderStatus.IsTerminal(), nil
	}

	if terminal, err := send(); err != nil || terminal {
		return err
	}

	ticker := time.NewTicker(poolInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if terminal, err := send(); err != nil || terminal {
				return err
			}
		}
	}
}
