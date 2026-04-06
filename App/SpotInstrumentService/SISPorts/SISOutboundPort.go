package SISOutboundPort

import (
	"context"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
)

type SISOutboundPort interface {
	ViewMarketByID(ctx context.Context, marketId string) (*SISDomain.MarketDomain, error)
	// View chosen market by ID of this market

	ViewAllMarkets(ctx context.Context) ([]*SISDomain.MarketDomain, error)
	// View all markets as a list, witch depends on a role
}
