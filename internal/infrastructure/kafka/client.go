package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stannisl/image-processor/internal/config"
)

type Client struct {
	cfg    *config.Config
	log    *slog.Logger
	writer *kafka.Writer
}

// NewClient создает клиента кафки. Не тред сейф
func NewClient(cfg *config.Config, log *slog.Logger) (*Client, error) {
	if log == nil {
		log = slog.Default()
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Kafka.Brokers...),
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: false,
		MaxAttempts:            cfg.Kafka.MaxRetries,
		Compression:            kafka.Lz4,
		Logger:                 kafka.LoggerFunc(func(s string, a ...any) { log.Debug(fmt.Sprintf(s, a...)) }),
		ErrorLogger:            kafka.LoggerFunc(func(s string, a ...any) { log.Error(fmt.Sprintf(s, a...)) }),
	}

	return &Client{
		writer: writer,
		log:    log,
		cfg:    cfg,
	}, nil
}

func (c *Client) Produce(ctx context.Context, topic, key string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("kafka_produce: json marshal: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}

	if err := c.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka produce write [topic=%s key=%s]: %w", topic, key, err)
	}

	c.log.Debug("kafka message produced", "topic", topic, "key", key, "bytes", len(payload))
	return nil
}

func (c *Client) NewReader(topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.cfg.Kafka.Brokers,
		Topic:          topic,
		GroupID:        groupID,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 0,
		Logger:         kafka.LoggerFunc(func(s string, a ...any) { c.log.Debug(fmt.Sprintf(s, a...)) }),
		ErrorLogger:    kafka.LoggerFunc(func(s string, a ...any) { c.log.Error(fmt.Sprintf(s, a...)) }),
		MinBytes:       1,
		MaxBytes:       10 << 20,
	})
}

func (c *Client) Close() error {
	return c.writer.Close()
}
