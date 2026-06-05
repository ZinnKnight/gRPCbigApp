package SISPorts

import (
	"context"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
)

type SISInboundPort interface {
	GetMarketByName(ctx context.Context, marketName string) (*SISDomain.MarketDomain, error)
	GetAllMarkets(ctx context.Context, pageSize int, curs string) ([]*SISDomain.MarketDomain, string, error)
}
