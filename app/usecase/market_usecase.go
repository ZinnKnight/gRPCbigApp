package usecase

import (
	"context"
	"gRPCbigapp/app/domain"
)

type MarketUseCaseImplementation struct {
	repo domain.MarketRepo
}

func NewMarketUseCaseImplementation(repo domain.MarketRepo) *MarketUseCaseImplementation {
	return &MarketUseCaseImplementation{repo: repo}
}

func (muci *MarketUseCaseImplementation) ViewMarket(ctx context.Context, userRole string) ([]domain.MarketDomain, error) {
	all, err := muci.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	active := make([]domain.MarketDomain, 0, len(all))

	for _, m := range all {
		if m.Accessibility && m.TTL == nil {
			active = append(active, m)
		}
	}
	return active, nil
}
