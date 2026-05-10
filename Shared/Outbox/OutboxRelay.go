package Outbox

import (
	"context"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	tracing "gRPCbigapp/Shared/Tracing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var relayTrancer = tracing.Tracer("outbox.relay")

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

	ctx, batchSpan := relayTrancer.Start(ctx, "outbox.realy.batching", tracing.KindInternal)
	defer batchSpan.End()

	events, err := r.repo.FetchUnpublished(ctx, r.batchSize)
	if err != nil {
		batchSpan.RecordError(err)
		batchSpan.SetStatus(codes.Error, "outbox, fetch increment retry fail")
		r.logger.LogError("Outbox FetchUnpublished failed",
			LoggerPorts.Fieled{Key: "error", Value: err.Error()},
		)
		return
	}
	batchSpan.SetAttributes(attribute.Int("batch.Size", len(events)))

	for _, event := range events {

		publishCtx := tracing.TakeOutFromCar(ctx, event.TraceContext)
		publishCtx, publishSpan := relayTrancer.Start(publishCtx, "outbox.publish", tracing.KindProducer)
		publishSpan.SetAttributes(tracing.OutboxMesseging(event.EventType)...)
		publishSpan.SetAttributes(
			attribute.String("messaging.messege.id", event.IdempotencyKey),
			attribute.String("aggregate.type", event.AggregatorType),
			attribute.String("aggregator.id", event.AggregatorID),
			attribute.Int64("outbox.event.id", event.ID),
		)

		if err := r.publisher.Publish(ctx, event); err != nil {
			publishSpan.RecordError(err)
			publishSpan.SetStatus(codes.Error, "publish failed")
			r.logger.LogError("Outbox Publish failed",
				LoggerPorts.Fieled{Key: "even_id", Value: event.ID},
				LoggerPorts.Fieled{Key: "event_type", Value: event.EventType},
				LoggerPorts.Fieled{Key: "error", Value: err.Error()},
			)
			_ = r.repo.IncrementRetry(ctx, event.ID)
			publishSpan.End()
			continue
		}

		if err := r.repo.MarkPublished(ctx, event.ID); err != nil {
			r.logger.LogError("Outbox MarkPublished failed",
				LoggerPorts.Fieled{Key: "even_id", Value: event.ID},
				LoggerPorts.Fieled{Key: "error", Value: err.Error()})
		}

		publishSpan.End()
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
