package repo

import (
	"context"
	"fmt"
	"gRPCbigapp/app/domain"
)

type RAMRepo struct {
	markets map[string]domain.MarketDomain
}

// это временное решение хранения данных в памяти, но озже можно спокойно переделать в построгвский адаптер

func NewRAMRepo() *RAMRepo {
	repo := &RAMRepo{
		markets: make(map[string]domain.MarketDomain),
	}
	repo.fixedOption()
	return repo
}

func (rr *RAMRepo) fixedOption() {
	rr.markets["Market1"] = domain.MarketDomain{
		MarketID:      "Market1",
		GoodsId:       "Goods1",
		Accessibility: true,
		TTL:           nil,
	}
	rr.markets["Market2"] = domain.MarketDomain{
		MarketID:      "Market2",
		GoodsId:       "Goods2",
		Accessibility: false,
		TTL:           nil,
	}
}

func (rr *RAMRepo) GetAll(_ context.Context) ([]domain.MarketDomain, error) {
	result := make([]domain.MarketDomain, 0, len(rr.markets))
	for _, market := range rr.markets {
		result = append(result, market)
	}
	return result, nil
}

func (rr *RAMRepo) GetByID(_ context.Context, id string) (*domain.MarketDomain, error) {
	market, ok := rr.markets[id]
	if !ok {
		return nil, fmt.Errorf("невозможно найти market: %s", id)
	}
	return &market, nil
}
