package acsessAdapter

import (
	"context"
	marketpb "gRPCbigapp/protofiles/gRPCbigapp/Protofiles/markets"

	"google.golang.org/grpc"
)

type GRPCMarketAdapter struct {
	client marketpb.SpotInstrumentServiceClient
}

func NewGRPCAccessAdapter(connection *grpc.ClientConn) *GRPCMarketAdapter {
	return &GRPCMarketAdapter{
		client: marketpb.NewSpotInstrumentServiceClient(connection)}
}

func (adapter *GRPCMarketAdapter) MarketAlive(ctx context.Context, marketId, userRole string) (bool, error) {
	resp, err := adapter.client.ViewMarkets(ctx, &marketpb.ViewMarketsRequest{
		UserRole: userRole,
	})
	if err != nil {
		return false, err
	}

	for _, m := range resp.Markets {
		if m.MarketId == marketId {
			return true, nil
		}
	}
	return false, nil
}
