package SISDomain

import (
	"errors"
	"time"
)

type MarketDomain struct {
	GoodsID       string
	MarketID      string
	Accessibility bool
	TTL           *time.Time
}

var (
	ErrMarketNotFound = errors.New("market not found")
)
