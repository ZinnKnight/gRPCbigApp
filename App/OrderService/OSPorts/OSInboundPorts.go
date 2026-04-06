package OSInboundPorts

import (
	"context"
	"gRPCbigapp/App/OrderService/OSDomain"
	"gRPCbigapp/App/OrderService/OSUseCase"
	"gRPCbigapp/App/Shared/Auth/AuthUseCase"
	"gRPCbigapp/App/Shared/Logger/LoggerPorts"
	orderpb "gRPCbigapp/Proto/order"
	"time"

	"github.com/google/uuid"
)

type OSInboundPorts struct {
	osiP   *OSUseCase.OrderService
	logger LoggerPorts.Logger
}

func NewOSLoggerService(log *LoggerPorts.Logger, osip *OSUseCase.OrderService) *OSInboundPorts {
	return &OSInboundPorts{osiP: osip, logger: *log}
}

func (osp *OSInboundPorts) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	user := AuthUseCase.GetUserFromContext(ctx)
	order := &OSDomain.OrderDomain{
		UserID:      user.UserID,
		OrderID:     uuid.New().String(),
		MarketName:  req.MarketId,
		Price:       req.Price,
		Amount:      req.Quantity,
		OrderStatus: "Created",
		CreatedAt:   time.Now(),
	}
	err := osp.osiP.CreateNewOrder(ctx, order)
	if err != nil {

		// TODO ROLLBACK or some type of that

		osp.logger.LogError("Error in CreateOrderMethod method",
			LoggerPorts.Fieled{Key: "user_id", Value: user.UserID},
			LoggerPorts.Fieled{Key: "Error", Value: err.Error()})
		return nil, err
	}
	return &orderpb.CreateOrderResponse{}, nil
	// TODO need to heck does it work like that
}

func (osp *OSInboundPorts) OrderStatusByID(ctx context.Context, req *orderpb.OrderStatusByIDRequest) (*orderpb.OrderStatusByIDResponse, error) {
	order, err := osp.osiP.GetOrderStatusById(ctx, req.OrderId)
	if err != nil {

		// TODO ROLLBACK or some type of that

		osp.logger.LogError("Error in OrderStatusByIDMethod method",
			LoggerPorts.Fieled{Key: "order_id", Value: order},
			LoggerPorts.Fieled{Key: "Error", Value: err.Error()})
		return nil, err
	}
	return &orderpb.OrderStatusByIDResponse{
		// TODO like that for now, but need to think how o make it work properly
	}, nil
}

func (osp *OSInboundPorts) OrderStatusAll(ctx context.Context, req *orderpb.OrderStatusAllRequest) (*orderpb.OrderStatusAllResponse, error) {
	orders, err := osp.osiP.GetOrderStatusAll(ctx)
	if err != nil {

		// TODO ROLLBACK or some type of that

		osp.logger.LogError("Error in OrderStatusAllMethod method",
			LoggerPorts.Fieled{Key: "orders_id, order_status", Value: orders},
			LoggerPorts.Fieled{Key: "Error", Value: err.Error()})
		return nil, err
	}
	return &orderpb.OrderStatusAllResponse{
		// TODO like that for now, but need to think how o make it work properly
	}, nil
}

func (osp *OSInboundPorts) StreamOrderStatus(ctx context.Context, req *orderpb.StreamOrderRequest) (*orderpb.OrderStatusByIDResponse, error) {
	streamOrder, err := osp.osiP.StreamOrderStatus(ctx, req.OrderId)
	if err != nil {

		// TODO ROLLBACK or some type of that

		osp.logger.LogError("Error in StreamOrderStatusMethod method",
			LoggerPorts.Fieled{Key: "order_id, order_status", Value: streamOrder},
			LoggerPorts.Fieled{Key: "Error", Value: err.Error()})
		return nil, err
	}
	return &orderpb.OrderStatusByIDResponse{
		// TODO like that for now, but need to think how o make it work properly
	}, nil
}
