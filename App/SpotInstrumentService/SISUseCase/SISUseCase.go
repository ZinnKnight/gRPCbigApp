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

func (sis *SISUseCase) GetAllMarkets(ctx context.Context, pageSize int, curs string) ([]*SISDomain.MarketDomain, string, error) {
	markets, err := sis.repo.FindAll(ctx, pageSize+1, curs)
	if err != nil {
		return nil, "", fmt.Errorf("usecase, failed to get all markets: %w", err)
	}

	var next string

	if len(markets) > pageSize {
		next = markets[pageSize-1].MarketID
		markets = markets[:pageSize]
	}
	return markets, next, nil
}
