package grpcAdapters

import (
	"context"
	"gRPCbigapp/App/domain"
	"gRPCbigapp/App/mapper"
	marketpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/markets"
)

type MarketService struct {
	marketpb.UnimplementedSpotInstrumentServiceServer
	markets map[string]domain.MarketDomain
}

func NewMarketService() *MarketService {
	return &MarketService{markets: make(map[string]domain.MarketDomain)}
}

func (ms *MarketService) ViewMarkets(ctx context.Context, req *marketpb.ViewMarketsRequest) (*marketpb.ViewMarketsResponse, error) {
	resp := &marketpb.ViewMarketsResponse{}

	for _, market := range ms.markets {
		if market.Accessibility {
			resp.Markets = append(resp.Markets, mapper.MapperMarketsToProto(market))
		}
	}
	return resp, nil
}
