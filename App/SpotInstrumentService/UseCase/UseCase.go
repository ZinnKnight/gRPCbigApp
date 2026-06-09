package UseCase

import (
	"context"
	"fmt"
	"gRPCbigapp/App/SpotInstrumentService/Domain"
	"gRPCbigapp/App/SpotInstrumentService/Ports"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	tracing "gRPCbigapp/Shared/Tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var trace = tracing.Tracer("usecase.UseCase")

var _ Ports.SISInboundPort = (*SISUseCase)(nil)

type SISUseCase struct {
	repo   Ports.SISOutboundRepo
	logger LoggerPorts.Logger
}

func NewSISUseCase(repo Ports.SISOutboundRepo, logger LoggerPorts.Logger) *SISUseCase {
	return &SISUseCase{repo: repo, logger: logger}
}

func (sis *SISUseCase) GetMarketByName(ctx context.Context, marketName string) (*Domain.MarketDomain, error) {

	ctx, span := trace.Start(ctx, "GetMarketByID", tracing.KindClient)
	defer span.End()

	span.SetAttributes(attribute.String("marketID", marketName))

	market, err := sis.repo.FindByName(ctx, marketName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repository.GetMarketByID failed")
		return nil, fmt.Errorf("usecase, failed to get market: %w", err)
	}
	return market, nil
}

func (sis *SISUseCase) GetAllMarkets(ctx context.Context, pageSize int, curs string) ([]*Domain.MarketDomain, string, error) {

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
