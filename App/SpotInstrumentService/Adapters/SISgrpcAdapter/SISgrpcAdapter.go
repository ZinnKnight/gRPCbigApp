package SISgrpcAdapter

import (
	"context"
	"gRPCbigapp/App/SpotInstrumentService/SISDomain"
	"gRPCbigapp/App/SpotInstrumentService/SISPorts"
	marketpb "gRPCbigapp/Proto/protoPB/marketPB"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
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

func protoMarketMapper(m *SISDomain.MarketDomain) *marketpb.Market {
	var ttl int64

	if m.TTL != nil {
		ttl = m.TTL.Unix()
	}

	return &marketpb.Market{
		MarketName:          m.MarketName,
		GoodsId:             m.GoodsID,
		MarketId:            m.MarketID,
		MarketAccessibility: m.Accessibility,
		MarketTtl:           ttl,
	}

}

func (h *SISgrpcHandler) ViewMarketByID(ctx context.Context, req *marketpb.ViewMarketRequest) (*marketpb.ViewMarketResponse, error) {
	market, err := h.useCase.GetMarketByName(ctx, req.GetViewMarketRequest())
	if err != nil {
		h.logger.LogError("sisgrpcAdapter, failed to view a market by id",
			LoggerPorts.Field{Key: "name", Value: req.GetViewMarketRequest()},
			LoggerPorts.Field{Key: "error", Value: err.Error()})
		return nil, err
	}
	return &marketpb.ViewMarketResponse{
		ViewMarketResponse: protoMarketMapper(market),
	}, nil
}

func (h *SISgrpcHandler) ViewAllMarkets(ctx context.Context, req *marketpb.ViewMarketsAllRequest) (*marketpb.ViewMarketsAllResponse, error) {
	size := int(req.GetPageSize())
	if size <= 0 || size > 50 {
		size = 10
	}

	markets, nextCurs, err := h.useCase.GetAllMarkets(ctx, size, req.GetPageToken())
	if err != nil {
		h.logger.LogError("sisgrpcAdapter, failed to view all markets",
			LoggerPorts.Field{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	pbMarkets := make([]*marketpb.Market, 0, len(markets))

	for _, market := range markets {
		pbMarkets = append(pbMarkets, protoMarketMapper(market))
	}
	return &marketpb.ViewMarketsAllResponse{Markets: pbMarkets, NextPageToken: nextCurs}, nil
}
