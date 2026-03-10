package tests

import (
	"context"
	marketTest "gRPCbigapp/App/adapters/grpcAdapters"
	marketpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/markets"
	"testing"
)

func TestViewMarketsReturns(t *testing.T) {
	marketTest.NewMarketService()

	service := marketTest.NewMarketService()

	response, err := service.ViewMarkets(
		context.Background(),
		&marketpb.ViewMarketsRequest{},
	)

	if err != nil {
		t.Fatalf("Непредвиденное поведение: %v", err)
	}

	for _, m := range response.Markets {
		if !m.Enable {
			t.Errorf("Данный маркет недоступен: %v", err)
		}
	}

}
