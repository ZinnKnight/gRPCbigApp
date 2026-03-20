package domain

import "time"

type MarketDomain struct {
	GoodsId       string
	MarketID      string
	Accessibility bool
	TTL           *time.Time
}
