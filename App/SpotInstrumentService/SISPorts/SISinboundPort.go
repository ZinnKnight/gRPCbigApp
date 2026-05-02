package SISPorts

import (
	"context"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
)

type SISInboundPort interface {
	GetMarketByID(ctx context.Context, marketID string) (*SISDomain.MarketDomain, error)
	GetAllMarkets(ctx context.Context, pageSize int, curs string) ([]*SISDomain.MarketDomain, string, error)
}
