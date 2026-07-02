package Kafka

import (
	"context"
	"fmt"
	"gRPCbigapp/Shared/Logger/LoggerPorts"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Message struct {
	Topic  string
	Key    []byte
	Value  []byte
	Header map[string]string
}

type Handler func(ctx context.Context, message Message) error

type ConsumerConfig struct {
	Brokers    []string
	Group      string
	Topics     []string
	StartAtEnd bool
}

type Consumer struct {
	client *kgo.Client
	logger LoggerPorts.Logger
}

func NewConsumer(ctx context.Context, config ConsumerConfig, logger LoggerPorts.Logger) (*Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(config.Brokers...),
		kgo.ConsumerGroup(config.Group),
		kgo.ConsumeTopics(config.Topics...),
		kgo.DisableAutoCommit(),
	}
	if config.StartAtEnd {
		opts = append(opts, kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()))
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer, new client: %w", err)
	}

	pingCTX, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := cl.Ping(pingCTX); err != nil {
		cl.Close()
		return nil, fmt.Errorf("kafka consumer, ping client: %w", err)
	}

	return &Consumer{
		client: cl,
		logger: logger,
	}, nil
}

func (c *Consumer) Run(ctx context.Context, handler Handler) {
	c.logger.LogInfo("kafka consumer started")
	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			c.logger.LogInfo("kafka consumer stopped (client close)")
			return
		}
		if ctx.Err() != nil {
			c.logger.LogInfo("kafka consumer stopped (ctx done)")
			return
		}

		fetches.EachError(func(topic string, partition int32, err error) {
			c.logger.LogError("kafka consumer fetch error",
				LoggerPorts.Field{Key: "topic", Value: topic},
				LoggerPorts.Field{Key: "partition", Value: partition},
				LoggerPorts.Field{Key: "error", Value: err.Error()},
			)
		})

		var ok []*kgo.Record

		iter := fetches.RecordIter()
		for !iter.Done() {
			rec := iter.Next()
			headers := make(map[string]string, len(rec.Headers))
			for _, head := range rec.Headers {
				headers[head.Key] = string(head.Value)
			}
			err := handler(ctx, Message{
				Topic:  rec.Topic,
				Key:    rec.Key,
				Value:  rec.Value,
				Header: headers,
			})
			if err != nil {
				c.logger.LogError("kafka consumer handler error",
					LoggerPorts.Field{Key: "topic", Value: rec.Topic},
					LoggerPorts.Field{Key: "error", Value: err.Error()})
				break
			}
			ok = append(ok, rec)
		}

		if len(ok) > 0 {
			if err := c.client.CommitRecords(ctx, ok...); err != nil {
				c.logger.LogError("kafka consumer commit error",
					LoggerPorts.Field{Key: "error", Value: err.Error()})
			}
		}
	}
}

func (c *Consumer) Close() {
	c.client.Close()
}
