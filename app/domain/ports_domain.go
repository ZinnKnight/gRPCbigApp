package domain

import "context"

// эти 2 структуры я оставил тут т.к не знал куда ещё их пристроить, по идее можно их и в отдельный domain файлик вынести,
// но зачем с другой стороны, тип не будет ли это оверхедом

type CreateOrderParameters struct {
	UserId   string
	MarketId string
	UserRole string
	Price    float64
	Quantity float64
}

type OrderResult struct {
	OrderId     string
	OrderStatus string
}

type OrderRepo interface {
	Create(ctx context.Context, od *OrderDomain) error
	GetStatus(ctx context.Context, id string) (string, error)
}

type MarketRepo interface {
	GetAll(ctx context.Context) ([]MarketDomain, error)
	GetByID(ctx context.Context, marketId string) (*MarketDomain, error)
}

type MarketAcessChecker interface {
	MarketAlive(ctx context.Context, marketId, userRole string) (bool, error)
}

type OrderUsecase interface {
	CreateOrder(ctx context.Context, parameters CreateOrderParameters) (*OrderResult, error)
	GetOrderStatus(ctx context.Context, orderId string) (string, error)
}

type MarketUsecase interface {
	ViewMarket(ctx context.Context, userRole string) ([]MarketDomain, error)
}
