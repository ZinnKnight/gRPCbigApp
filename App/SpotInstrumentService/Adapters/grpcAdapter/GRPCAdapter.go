package grpcAdapter

import (
	"context"
	"gRPCbigapp/App/SpotInstrumentService/Domain"
	"gRPCbigapp/App/SpotInstrumentService/Ports"
	"gRPCbigapp/Proto/protoPB"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
)

type SISgrpcHandler struct {
	protoPB.UnimplementedSpotInstrumentServiceServer
	useCase Ports.SISInboundPort
	logger  LoggerPorts.Logger
}

func NewSISgrpcHandler(sisH Ports.SISInboundPort, logger LoggerPorts.Logger) *SISgrpcHandler {
	return &SISgrpcHandler{
		useCase: sisH,
		logger:  logger,
	}
}

func protoMarketMapper(m *Domain.MarketDomain) *protoPB.Market {
	var ttl int64

	if m.TTL != nil {
		ttl = m.TTL.Unix()
	}

	return &protoPB.Market{
		MarketName:          m.MarketName,
		GoodsId:             m.GoodsID,
		MarketId:            m.MarketID,
		MarketAccessibility: m.Accessibility,
		MarketTtl:           ttl,
	}

}

func (h *SISgrpcHandler) ViewMarketByID(ctx context.Context, req *protoPB.ViewMarketRequest) (*protoPB.ViewMarketResponse, error) {
	market, err := h.useCase.GetMarketByName(ctx, req.GetViewMarketRequest())
	if err != nil {
		h.logger.LogError("sisgrpcAdapter, failed to view a market by id",
			LoggerPorts.Field{Key: "name", Value: req.GetViewMarketRequest()},
			LoggerPorts.Field{Key: "error", Value: err.Error()})
		return nil, err
	}
	return &protoPB.ViewMarketResponse{
		ViewMarketResponse: protoMarketMapper(market),
	}, nil
}

func (h *SISgrpcHandler) ViewAllMarkets(ctx context.Context, req *protoPB.ViewMarketsAllRequest) (*protoPB.ViewMarketsAllResponse, error) {
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

	pbMarkets := make([]*protoPB.Market, 0, len(markets))

	for _, market := range markets {
		pbMarkets = append(pbMarkets, protoMarketMapper(market))
	}
	return &protoPB.ViewMarketsAllResponse{Markets: pbMarkets, NextPageToken: nextCurs}, nil
}
