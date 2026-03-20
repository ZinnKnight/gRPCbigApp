package usecase

import (
	"context"
	"fmt"
	"gRPCbigapp/app/domain"
	"time"

	"github.com/google/uuid"
)

type OrderUseCaseImplementation struct {
	repo        domain.OrderRepo
	marketAcess domain.MarketAcessChecker
}

func NewOrderUseCaseImplementation(repo domain.OrderRepo, marketAcess domain.MarketAcessChecker) *OrderUseCaseImplementation {
	return &OrderUseCaseImplementation{
		repo:        repo,
		marketAcess: marketAcess,
	}
}

func (ouci *OrderUseCaseImplementation) CreateOrder(ctx context.Context, params domain.CreateOrderParameters) (*domain.OrderResult, error) {
	if params.UserId == "" || params.MarketId == "" {
		return nil, fmt.Errorf("userId или marketId не могут быть пустыми")
	}
	if params.Price < 0 || params.Quantity < 0 {
		return nil, fmt.Errorf("price или quantity не могут отрицательными")
	}

	if ouci.marketAcess != nil {
		active, err := ouci.marketAcess.MarketAlive(ctx, params.MarketId, params.UserId)
		if err != nil {
			return nil, fmt.Errorf("невозможно получить информацию о market: %w", err)
		}
		if !active {
			return nil, fmt.Errorf("не удаётся получить информацию о работе market, сервис недоступен или удалён: %w", err)
		}
	}
	order := &domain.OrderDomain{
		UserID:      params.UserId,
		OrderID:     uuid.New().String(),
		MarketName:  params.MarketId,
		Price:       params.Price,
		Amount:      params.Quantity,
		OrderStatus: "Created",
		CreatedAt:   time.Now(),
	}

	if err := ouci.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("невозможно создать заказ: %w", err)
	}

	return &domain.OrderResult{
		OrderId:     order.OrderID,
		OrderStatus: order.OrderStatus,
	}, nil
}

func (ouci *OrderUseCaseImplementation) GetOrderStatus(ctx context.Context, orderId string) (string, error) {
	if orderId == "" {
		return "", fmt.Errorf("для получения статуса необходим orderId")
	}
	return ouci.repo.GetStatus(ctx, orderId)
}
