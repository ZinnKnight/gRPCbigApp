package SISInboundPort

import (
	"context"
	"gRPCbigapp/App/Shared/Logger/LoggerPorts"
	"gRPCbigapp/App/SpotInstrumentService/SISUseCase"
	sispb "gRPCbigapp/Proto/market"
)

type SISInboundHandler struct {
	sisIH  *SISUseCase.SpotInstrumentService
	logger LoggerPorts.Logger
}

func NewSISLoggerService(log LoggerPorts.Logger, sisih *SISUseCase.SpotInstrumentService) *SISInboundHandler {
	return &SISInboundHandler{logger: log, sisIH: sisih}
}

func (h *SISInboundHandler) ViewMarketByID(ctx context.Context, req *sispb.ViewMarketsByIDRequest) (*sispb.ViewMarketsByIDResponse, error) {
	market, err := h.sisIH.ViewMarket(ctx, req.MarketId)
	if err != nil {

		// TODO ROLLBACK or some type of that

		h.logger.LogError("Error in ViewMarketByID method",
			LoggerPorts.Fieled{Key: "market_id", Value: market},
			LoggerPorts.Fieled{Key: "Error", Value: err.Error()})
		return nil, err
	}
	return &sispb.ViewMarketsByIDResponse{
		// TODO we return something
	}, nil
}

func (h *SISInboundHandler) ViewAllMarkets(ctx context.Context, req *sispb.ViewMarketsAllRequest) (*sispb.ViewMarketsAllResponse, error) {
	allMarkets, err := h.sisIH.ViewAllMarkets(ctx)
	if err != nil {

		// TODO ROLLBACK or some type of that

		h.logger.LogError("Error in ViewAllMarkets method",
			LoggerPorts.Fieled{Key: "markets", Value: allMarkets},
			LoggerPorts.Fieled{Key: "Error", Value: err.Error()})
		return nil, err
	}
	return &sispb.ViewMarketsAllResponse{
		// TODO we return something
	}, nil
}
