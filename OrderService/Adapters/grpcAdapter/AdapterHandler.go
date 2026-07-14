package grpcAdapter

import (
	"context"
	"gRPCbigapp/ClientService/Auth/AuthCTX"
	"gRPCbigapp/OrderService/Domain"
	"gRPCbigapp/OrderService/Ports"
	"gRPCbigapp/OrderService/Streaming"
	"gRPCbigapp/Proto/protoPB"
	moneyconverter "gRPCbigapp/Shared/Converters/Money"
	"gRPCbigapp/Shared/ErrorInterceptor"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
)

// todo пересмотреть, сейчас не гексагоналка

type OrderHandler struct {
	protoPB.UnimplementedOrderServiceServer
	useCase Ports.OSInboundPort
	hub     *Streaming.Hub
	logger  LoggerPorts.Logger
}

func NewOrderHandler(log LoggerPorts.Logger, osp Ports.OSInboundPort, hub *Streaming.Hub) *OrderHandler {
	return &OrderHandler{
		logger:  log,
		useCase: osp,
		hub:     hub,
	}
}

// сразу просматриваем статусы/конвертеры из pb.proto

func pbProtoStatuses(status Domain.OrderStatus) protoPB.OrderStatus {
	if val, ok := protoPB.OrderStatus_value[string(status)]; ok {
		return protoPB.OrderStatus(val)
	}
	return protoPB.OrderStatus_UNREGISTERED_STATUS
}

func pbOrderConverter(convert *Domain.OrderDomain) *protoPB.Order {
	return &protoPB.Order{
		UserId:      convert.UserID,
		OrderId:     convert.OrderID,
		MarketId:    convert.MarketID,
		Price:       moneyconverter.DecToMoney(convert.Price, moneyconverter.Currency),
		Amount:      moneyconverter.DecimalToDecimalPB(convert.Amount),
		OrderStatus: pbProtoStatuses(convert.OrderStatus),
		CreatedAt:   convert.CreatedAt.Unix(),
	}
}

func (o *OrderHandler) CreateOrder(ctx context.Context, req *protoPB.CreateOrderRequest) (*protoPB.CreateOrderResponse, error) {
	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}

	price := moneyconverter.MoneyToDec(req.GetPrice())

	amount, err := moneyconverter.DecimalPBToDecimal(req.GetAmount())
	if err != nil {
		return nil, ErrorInterceptor.NewError(ErrorInterceptor.Invalid, "Некорректная сумма заказа", err)
	}

	cmd := Ports.CreteOrder{
		UserID:   user.UserID,
		MarketID: req.GetMarketId(),
		Price:    price,
		Quantity: amount,
		UserPlan: user.UserPlan,
	}

	orderID, err := o.useCase.CreateOrder(ctx, cmd)
	if err != nil {
		o.logger.LogError("grpc, failed to crete order",
			LoggerPorts.Field{Key: "id", Value: user.UserID},
			LoggerPorts.Field{Key: "error", Value: err.Error()})
		return nil, err
	}

	return &protoPB.CreateOrderResponse{CreateOrderResponse: pbOrderConverter(orderID)}, nil
}

func (o *OrderHandler) GetOrderStatusByID(ctx context.Context, req *protoPB.OrderStatusByIDRequest) (*protoPB.OrderStatusByIDResponse, error) {
	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}
	order, err := o.useCase.GetOrderStatusByID(ctx, req.GetOrderId(), user.UserID)
	if err != nil {
		return nil, err
	}
	return &protoPB.OrderStatusByIDResponse{
		OrderStatusResponse: pbOrderConverter(order),
	}, nil
}

func (o *OrderHandler) GetOrderStatusAll(ctx context.Context, req *protoPB.OrderStatusAllRequest) (*protoPB.OrderStatusAllResponse, error) {
	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}

	size := int(req.GetPageSize())
	if size <= 0 || size > 100 {
		size = 20
	}

	orders, nextPageToken, err := o.useCase.GetOrderStatusAll(ctx, user.UserID, req.GetPageToken(), size)
	if err != nil {
		return nil, err
	}

	protoOrders := make([]*protoPB.Order, 0, len(orders))

	for _, ord := range orders {
		protoOrders = append(protoOrders, pbOrderConverter(ord))
	}

	return &protoPB.OrderStatusAllResponse{
		AllOrdersStatusesResponse: protoOrders,
		NextPageToken:             nextPageToken,
	}, nil
}

func (o *OrderHandler) StreamOrderUpdates(req *protoPB.StreamOrderRequest, stream protoPB.OrderService_StreamOrderUpdatesServer) error {
	ctx := stream.Context()

	user, ok := AuthCTX.GetUser(ctx)
	if !ok {
		return ErrUnauthenticated
	}

	orderID := req.GetOrderId()

	subID, updates := o.hub.Subscribe(orderID)
	defer o.hub.Unsubscribe(orderID, subID)

	var lastSend Domain.OrderStatus
	firstSend := true

	send := func() (terminal bool, err error) {
		order, err := o.useCase.GetOrderStatusByID(ctx, orderID, user.UserID)

		if err != nil {
			return false, ErrorInterceptor.GRPCConnector(err)
		}

		if firstSend || order.OrderStatus != lastSend {
			if err := stream.Send(&protoPB.OrderStatusByIDResponse{
				OrderStatusResponse: pbOrderConverter(order)}); err != nil {
				// пока оставил так, позже буду мапить ошибки через отдельный обработчик
				return false, err
			}
			lastSend = order.OrderStatus
			firstSend = false
		}
		return order.OrderStatus.IsTerminal(), nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-updates:
			if terminal, err := send(); err != nil || terminal {
				return err
			}
		}
	}
}
