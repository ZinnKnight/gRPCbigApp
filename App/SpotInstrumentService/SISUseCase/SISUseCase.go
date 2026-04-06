package SISUseCase

import (
	"context"
	"gRPCbigapp/SpotInstrumentService/SISDomain"
	"gRPCbigapp/SpotInstrumentService/SISPorts/SISOutboundPort"
)

type SpotInstrumentService struct {
	repo SISOutboundPort.SISOutboundPort
}

func NewSpotInstrumentService(r SISOutboundPort.SISOutboundPort) *SpotInstrumentService {
	return &SpotInstrumentService{repo: r}
}

func (sis *SpotInstrumentService) ViewMarket(ctx context.Context, marketID string) (*SISDomain.MarketDomain, error) {
	return sis.repo.ViewMarketByID(ctx, marketID)
}

func (sis *SpotInstrumentService) ViewAllMarkets(ctx context.Context) (*[]SISDomain.MarketDomain, error) {
	return sis.repo.ViewAllMarkets(ctx)
	// TODO need to check how to return slice of it
}
