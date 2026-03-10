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

	markets := map[string]domain.MarketDomain{
		"Некий_Товар1": {GoodsId: "НекийТовар1", MarketID: "Некий-Ecomerce1", Accessibility: true},
		"Некий_Товар2": {GoodsId: "НекийТовар2", MarketID: "Некий-Ecomerce2", Accessibility: false},
		"Некий_Товар3": {GoodsId: "НекийТовар3", MarketID: "Некий-Ecomerce3", Accessibility: true},
	}
	return &MarketService{markets: markets}
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
