package SISUseCase

import (
	"context"
	"fmt"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
	"gRPCbigapp/App/SpotInstrumentService/SISPorts"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
)

var _ SISPorts.SISInboundPort = (*SISUseCase)(nil)

type SISUseCase struct {
	repo   SISPorts.SISOutboundRepo
	logger LoggerPorts.Logger
}

func NewSISUseCase(repo SISPorts.SISOutboundRepo, logger LoggerPorts.Logger) *SISUseCase {
	return &SISUseCase{repo: repo, logger: logger}
}

func (sis *SISUseCase) GetMarketByID(ctx context.Context, marketID string) (*SISDomain.MarketDomain, error) {
	market, err := sis.repo.FindByID(ctx, marketID)
	if err != nil {
		return nil, fmt.Errorf("usecase, failed to get market: %w", err)
	}
	return market, nil
}

func (sis *SISUseCase) GetAllMarkets(ctx context.Context) ([]*SISDomain.MarketDomain, error) {
	markets, err := sis.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase, failed to get all markets: %w", err)
	}
	return markets, nil
}
