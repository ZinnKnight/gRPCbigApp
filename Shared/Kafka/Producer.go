package Kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Config struct {
	Brokers  []string
	ClientID string
}

type Header struct {
	Key   string
	Value []byte
}

type Producer struct {
	client *kgo.Client
}

func NewProducer(ctx context.Context, config Config) (*Producer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(config.Brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	}
	if config.ClientID != "" {
		opts = append(opts, kgo.ClientID(config.ClientID))
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("kafka, new client: %w", err)
	}

	pingCTX, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := cl.Ping(pingCTX); err != nil {
		cl.Close()
		return nil, fmt.Errorf("kafka, ping brokers: %w", err)
	}

	return &Producer{
		client: cl,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte, headers ...Header) error {
	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	for _, h := range headers {
		record.Headers = append(record.Headers, kgo.RecordHeader{
			Key:   h.Key,
			Value: h.Value,
		})
		if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
			return fmt.Errorf("kafka.ProduceSync to %q: %w", topic, err)
		}
	}
	return nil
}

func (p *Producer) Close() {
	p.client.Close()
}
