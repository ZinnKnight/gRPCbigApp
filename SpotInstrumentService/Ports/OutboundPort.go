package Ports

import (
	"context"
	"gRPCbigapp/SpotInstrumentService/Domain"
)

type SISOutboundRepo interface {
	FindByName(ctx context.Context, marketName string) (*Domain.MarketDomain, error)
	FindAll(ctx context.Context, lim int, curs string) ([]*Domain.MarketDomain, error)
}
