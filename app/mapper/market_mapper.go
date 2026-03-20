package mapper

import (
	"gRPCbigapp/app/domain"
	marketpb "gRPCbigapp/protofiles/gRPCbigapp/Protofiles/markets"
)

func DomainToProtoMarket(m domain.MarketDomain) *marketpb.Market {
	return &marketpb.Market{
		MarketId: m.MarketID,
		Enable:   m.Accessibility,
	}
}

func DomainToProtoMarkets(m []domain.MarketDomain) []*marketpb.Market {
	res := make([]*marketpb.Market, 0, len(m))
	for _, m := range m {
		res = append(res, DomainToProtoMarket(m))
	}
	return res
}
