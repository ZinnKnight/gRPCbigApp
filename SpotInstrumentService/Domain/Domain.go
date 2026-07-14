package Domain

import (
	"errors"
	"time"
)

type MarketDomain struct {
	MarketName    string
	GoodsID       string
	MarketID      string
	Accessibility bool
	TTL           *time.Time
}

// по хорошему в дальнейшем нужен будет отдельный домен под сагу, где будут собраны все интерфейсы, но пока что ошибки тут

const (
	ReservationReserved = "RESERVED"
	ReservationRejected = "REJECTED"
)

var (
	ErrMarketNotFound = errors.New("market not found")
)
