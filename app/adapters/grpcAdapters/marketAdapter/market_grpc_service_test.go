package marketAdapter

import (
	"context"
	marketpb "gRPCbigapp/protofiles/gRPCbigapp/Protofiles/markets"
	"testing"
)

func TestViewMarketsReturns(t *testing.T) {
	service := NewMarketService()

	response, err := service.ViewMarkets(
		context.Background(),
		&marketpb.ViewMarketsRequest{},
	)

	if err != nil {
		t.Fatalf("Непредвиденное поведение: %v", err)
	}

	for _, m := range response.Markets {
		if !m.Enable {
			t.Errorf("Данный маркет недоступен: %v", m)
		}
	}
}
