package SISgrpcAdapter

import (
	"context"
	"errors"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
	"gRPCbigapp/App/SpotInstrumentService/SISPorts"
	marketpb "gRPCbigapp/Proto/market"
	"gRPCbigapp/Shared/Logger/LoggerPorts"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SISgrpcHandler struct {
	marketpb.UnimplementedSpotInstrumentServiceServer
	useCase SISPorts.SISInboundPort
	logger  LoggerPorts.Logger
}

func NewSISgrpcHandler(sisH SISPorts.SISInboundPort, logger LoggerPorts.Logger) *SISgrpcHandler {
	return &SISgrpcHandler{
		useCase: sisH,
		logger:  logger,
	}
}

func marketErrorsMaper(err error) error {
	switch {
	case errors.Is(err, SISDomain.ErrMarketNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "sisgrpcAdapter, Internal error")
	}
}

func (h *SISgrpcHandler) ViewMarketByID(ctx context.Context, req *marketpb.ViewMarketsByIDRequest) (*marketpb.ViewMarketsByIDResponse, error) {
	market, err := h.useCase.GetMarketByID(ctx, req.MarketId)
	if err != nil {
		h.logger.LogError("sisgrpcAdapter, failed to view a market by id",
			LoggerPorts.Fieled{Key: "id", Value: req.MarketId},
			LoggerPorts.Fieled{Key: "error", Value: err.Error()})
		return nil, marketErrorsMaper(err)
	}
	return &marketpb.ViewMarketsByIDResponse{
		Market: &marketpb.Market{
			MarketId: req.MarketId,
			Enable:   market.Accessibility,
		},
	}, nil
}

func (h *SISgrpcHandler) ViewAllMarkets(ctx context.Context, req *marketpb.ViewMarketsAllRequest) (*marketpb.ViewMarketsAllResponse, error) {
	size := int(req.PageSize)
	if size <= 0 || size > 50 {
		size = 10
	}

	curs := req.PageToken

	markets, nextCurs, err := h.useCase.GetAllMarkets(ctx, size, curs)
	if err != nil {
		h.logger.LogError("sisgrpcAdapter, failed to view all markets",
			LoggerPorts.Fieled{Key: "error", Value: err.Error()},
		)
		return nil, marketErrorsMaper(err)
	}

	pbMarkets := make([]*marketpb.Market, 0, len(markets))

	for _, market := range markets {
		pbMarkets = append(pbMarkets, &marketpb.Market{
			MarketId: market.MarketID,
			Enable:   market.Accessibility,
		})
	}
	return &marketpb.ViewMarketsAllResponse{Markets: pbMarkets, NextPageToken: nextCurs}, nil
}
