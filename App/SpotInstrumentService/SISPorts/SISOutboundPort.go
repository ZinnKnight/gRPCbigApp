package SISPorts

import (
	"context"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
)

type SISOutboundRepo interface {
	FindByName(ctx context.Context, marketName string) (*SISDomain.MarketDomain, error)
	FindAll(ctx context.Context, lim int, curs string) ([]*SISDomain.MarketDomain, error)
}
