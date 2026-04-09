package Outbox

import (
	"context"
	"gRPCbigapp/App/Shared/Logger/LoggerPorts"
	"time"
)

type Publisher interface {
	Publish(ctx context.Context, event *Event) error
}

type Relay struct {
	repo      *Repository
	publisher Publisher
	logger    LoggerPorts.Logger
	interval  time.Duration
	batchSize int
}

func NewRelay(repo *Repository, publisher Publisher, logger LoggerPorts.Logger, interval time.Duration, batchSize int) *Relay {
	return &Relay{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		interval:  interval,
		batchSize: batchSize,
	}
}

func (r *Relay) batching(ctx context.Context) {
	events, err := r.repo.FetchUnpublished(ctx, r.batchSize)
	if err != nil {
		r.logger.LogError("Outbox FetchUnpublished failed",
			LoggerPorts.Fieled{Key: "error", Value: err.Error()},
		)
		return
	}

	for _, event := range events {
		if err := r.publisher.Publish(ctx, event); err != nil {
			r.logger.LogError("Outbox Publish failed",
				LoggerPorts.Fieled{Key: "even_id", Value: event.ID},
				LoggerPorts.Fieled{Key: "event_type", Value: event.EventType},
				LoggerPorts.Fieled{Key: "error", Value: err.Error()},
			)
			_ = r.repo.IncrementRetry(ctx, event.ID)
		}

		if err := r.repo.MarkPublished(ctx, event.ID); err != nil {
			r.logger.LogError("Outbox MarkPublished failed",
				LoggerPorts.Fieled{Key: "even_id", Value: event.ID},
				LoggerPorts.Fieled{Key: "error", Value: err.Error()})
		}
	}
}

func (r *Relay) Start(ctx context.Context) {
	r.logger.LogInfo("Starting outbox relay",
		LoggerPorts.Fieled{Key: "interval", Value: r.interval.String()},
		LoggerPorts.Fieled{Key: "batch_size", Value: r.batchSize},
	)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.LogInfo("Outbox relay stopped")
			return
		case <-ticker.C:
			r.batching(ctx)
		}
	}
}
