package marketAdapter

import (
	"context"
	"gRPCbigapp/app/domain"
	"gRPCbigapp/app/mapper"
	marketpb "gRPCbigapp/protofiles/gRPCbigapp/Protofiles/markets"
)

type MarketGrpcAdapter struct {
	marketpb.UnimplementedSpotInstrumentServiceServer
	usercase domain.MarketUsecase
}

func NewMarketGrpcAdapter(usrCase domain.MarketUsecase) *MarketGrpcAdapter {
	return &MarketGrpcAdapter{
		usercase: usrCase,
	}
}

func (ma *MarketGrpcAdapter) MarketShowcase(ctx context.Context, req *marketpb.ViewMarketsRequest) (*marketpb.ViewMarketsResponse, error) {
	markets, err := ma.usercase.ViewMarket(ctx, req.UserRole)
	if err != nil {
		return nil, err
	}
	return &marketpb.ViewMarketsResponse{
		Markets: mapper.DomainToProtoMarkets(markets),
	}, nil
}
