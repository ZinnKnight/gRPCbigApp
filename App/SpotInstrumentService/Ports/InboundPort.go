package Ports

import (
	"context"
	"gRPCbigapp/App/SpotInstrumentService/Domain"
)

type SISInboundPort interface {
	ViewMarketsByID(ctx context.Context, marketName string) (*Domain.MarketDomain, error)
	ViewMarketsAll(ctx context.Context, pageSize int, curs string) ([]*Domain.MarketDomain, string, error)
}
