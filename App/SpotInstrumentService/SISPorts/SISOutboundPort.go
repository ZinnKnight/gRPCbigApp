package SISPorts

import (
	"context"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
)

type SISOutboundRepo interface {
	FindByID(ctx context.Context, marketID string) (*SISDomain.MarketDomain, error)
	FindAll(ctx context.Context, lim int, curs string) ([]*SISDomain.MarketDomain, error)
}
