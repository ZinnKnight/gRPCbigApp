package OSDomain

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

type OrderStatus string

// Add StatusRejected for more clarity with StatusCanceled
const (
	StatusCreated   OrderStatus = "order created"
	StatusCancelled OrderStatus = "order cancelled"
	StatusPrepared  OrderStatus = "order prepared"
	StatusRejected  OrderStatus = "order rejected"
	StatusOrderDone OrderStatus = "order done"
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
)

func NewOrder(userID, marketID string, price, amount decimal.Decimal) (*OrderDomain, error) {
	// Probably switch better, not 100% sure for growth
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
