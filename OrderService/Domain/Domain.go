package Domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderDomain struct {
	UserID      string
	OrderID     string
	MarketID    string
	Price       decimal.Decimal
	Amount      decimal.Decimal
	OrderStatus OrderStatus
	CreatedAt   time.Time
}
type CreteOrder struct {
	UserID   string
	MarketID string
	Price    decimal.Decimal
	Quantity decimal.Decimal
	UserPlan string
}
type OrderStatus string

const (
	StatusUnregistered OrderStatus = "UNREGISTERED_STATUS"
	StatusCreated      OrderStatus = "ORDER_CREATED"
	StatusReserved     OrderStatus = "ORDER_RESERVED"
	StatusRejected     OrderStatus = "ORDER_REJECTED"
	StatusOrderDone    OrderStatus = "ORDER_DONE"
	StatusInDelivery   OrderStatus = "ORDER_IN_DELIVERY"
)

func (s OrderStatus) IsTerminal() bool {
	return s == StatusOrderDone || s == StatusRejected
}

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidPrice       = errors.New("invalid price")
	ErrInvalidAmount      = errors.New("invalid amount")
	ErrInvalidMarketID    = errors.New("invalid order id")
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrOrderQuotaExceeded = errors.New("order status invalid")
)

func NewOrder(userID, marketID string, price, amount decimal.Decimal) (*OrderDomain, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if marketID == "" {
		return nil, ErrInvalidMarketID
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}
	if price.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidPrice
	}

	return &OrderDomain{
		OrderID:     uuid.New().String(),
		UserID:      userID,
		MarketID:    marketID,
		Price:       price,
		Amount:      amount,
		OrderStatus: StatusCreated,
		CreatedAt:   time.Now(),
	}, nil
}
