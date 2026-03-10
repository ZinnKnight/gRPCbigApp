package mapper

import (
	"gRPCbigapp/App/domain"
	marketpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/markets"
)

func MapperMarketsToProto(m domain.MarketDomain) *marketpb.Market {
	return &marketpb.Market{
		MarketId: m.MarketID,
		Enable:   m.Accessibility,
	}
}
