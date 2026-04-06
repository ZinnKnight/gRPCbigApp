package SISDomain

import "time"

type MarketDomain struct {
	GoodsID       string
	MarketID      string
	Accessibility bool
	TTL           *time.Time
}
