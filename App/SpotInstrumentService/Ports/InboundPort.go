package Ports

import (
	"context"
	"gRPCbigapp/App/SpotInstrumentService/Domain"
)

type SISInboundPort interface {
	GetMarketByName(ctx context.Context, marketName string) (*Domain.MarketDomain, error)
	GetAllMarkets(ctx context.Context, pageSize int, curs string) ([]*Domain.MarketDomain, string, error)
}
