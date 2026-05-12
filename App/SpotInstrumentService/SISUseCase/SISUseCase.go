package SISUseCase

import (
	"context"
	"fmt"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
	"gRPCbigapp/App/SpotInstrumentService/SISPorts"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	tracing "gRPCbigapp/Shared/Tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var trace = tracing.Tracer("usecase.SISUseCase")

var _ SISPorts.SISInboundPort = (*SISUseCase)(nil)

type SISUseCase struct {
	repo   SISPorts.SISOutboundRepo
	logger LoggerPorts.Logger
}

func NewSISUseCase(repo SISPorts.SISOutboundRepo, logger LoggerPorts.Logger) *SISUseCase {
	return &SISUseCase{repo: repo, logger: logger}
}

func (sis *SISUseCase) GetMarketByID(ctx context.Context, marketID string) (*SISDomain.MarketDomain, error) {

	ctx, span := trace.Start(ctx, "GetMarketByID", tracing.KindClient)
	defer span.End()

	span.SetAttributes(attribute.String("marketID", marketID))

	market, err := sis.repo.FindByID(ctx, marketID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repository.GetMarketByID failed")
		return nil, fmt.Errorf("usecase, failed to get market: %w", err)
	}
	return market, nil
}

func (sis *SISUseCase) GetAllMarkets(ctx context.Context, pageSize int, curs string) ([]*SISDomain.MarketDomain, string, error) {

	ctx, span := trace.Start(ctx, "GetAllMarkets", tracing.KindClient)
	defer span.End()

	span.SetAttributes(attribute.Int("pageSize", pageSize))

	markets, err := sis.repo.FindAll(ctx, pageSize+1, curs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repository.GetAllMarkets failed")
		return nil, "", fmt.Errorf("usecase, failed to get all markets: %w", err)
	}

	var next string

	if len(markets) > pageSize {
		next = markets[pageSize-1].MarketID
		markets = markets[:pageSize]
	}

	span.SetAttributes(attribute.Int("markets.amount", len(markets)))
	return markets, next, nil
}
