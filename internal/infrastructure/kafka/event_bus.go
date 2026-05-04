package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type EventBus struct {
	client *Client
	log    *slog.Logger
}

func NewEventBus(client *Client, log *slog.Logger) *EventBus {
	if log == nil {
		log = slog.Default()
	}
	return &EventBus{client: client, log: log}
}

func (b *EventBus) Close() error {
	return b.client.Close()
}

func (b *EventBus) PublishUploaded(ctx context.Context, e ImageUploadedEvent) error {
	e.CreatedAt = time.Now()
	if err := b.client.Produce(ctx, TopicUploaded, e.ImageID, e); err != nil {
		return fmt.Errorf("PublishUploaded: %w", err)
	}
	b.log.Info("published", "topic", TopicUploaded, "image_id", e.ImageID, "ops", e.Operation)
	return nil
}

func (b *EventBus) PublishProcessed(ctx context.Context, e ImageProcessedEvent) error {
	e.ProcessedAt = time.Now()
	if err := b.client.Produce(ctx, TopicProcessed, e.ImageID, e); err != nil {
		return fmt.Errorf("PublishProcessed: %w", err)
	}
	b.log.Info("published", "topic", TopicProcessed, "image_id", e.ImageID, "ms", e.ProcessingTimeMs)
	return nil
}

func (b *EventBus) PublishFailed(ctx context.Context, e ImageFailedEvent) error {
	e.FailedAt = time.Now()
	if err := b.client.Produce(ctx, TopicFailed, e.ImageID, e); err != nil {
		return fmt.Errorf("PublishFailed: %w", err)
	}
	b.log.Warn("published to DLQ", "topic", TopicFailed, "image_id", e.ImageID, "err", e.Error)
	return nil
}

// SubscribeUploaded запускает concurrency горутин, каждая читает из image.uploaded.
// Несколько экземпляров сервиса с одним groupID → Kafka сама распределит партиции.
//
// handler должен быть idempotent: при панике/ошибке сообщение будет передано повторно
// (до maxRetries раз), после чего уходит в DLQ.
func (b *EventBus) SubscribeUploaded(
	ctx context.Context,
	groupID string,
	workers int,
	maxRetries int,
	handler func(context.Context, ImageUploadedEvent) error,
) {
	for i := range workers {
		go func(workerID int) {
			r := b.client.NewReader(TopicUploaded, groupID)
			defer r.Close()

			b.log.Info("worker started", "topic", TopicUploaded, "group", groupID, "worker", workerID)

			b.consume(ctx, r, maxRetries,
				func(ctx context.Context, msg kafka.Message) error {
					var e ImageUploadedEvent
					if err := json.Unmarshal(msg.Value, &e); err != nil {
						return fmt.Errorf("unmarshal: %w", err)
					}
					return handler(ctx, e)
				},
				func(ctx context.Context, msg kafka.Message, lastErr error, attempt int) {
					var orig ImageUploadedEvent
					_ = json.Unmarshal(msg.Value, &orig)
					_ = b.PublishFailed(ctx, ImageFailedEvent{
						ImageID:       orig.ImageID,
						Error:         lastErr.Error(),
						OriginalEvent: orig,
						Attempt:       attempt,
					})
				})
		}(i)
	}
}

// SubscribeProcessed слушает image.processed (статус → БД / SSE-нотификации).
func (b *EventBus) SubscribeProcessed(
	ctx context.Context,
	groupID string,
	handler func(context.Context, ImageProcessedEvent) error,
) {
	go func() {
		r := b.client.NewReader(TopicProcessed, groupID)
		defer r.Close()

		b.consume(ctx, r, 3, func(ctx context.Context, msg kafka.Message) error {
			var e ImageProcessedEvent
			if err := json.Unmarshal(msg.Value, &e); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			return handler(ctx, e)
		}, nil)
	}()
}

// SubscribeFailed слушает DLQ image.failed (алёрты / мониторинг).
func (b *EventBus) SubscribeFailed(
	ctx context.Context,
	groupID string,
	handler func(context.Context, ImageFailedEvent) error,
) {
	go func() {
		r := b.client.NewReader(TopicFailed, groupID)
		defer r.Close()

		b.consume(ctx, r, 1, func(ctx context.Context, msg kafka.Message) error {
			var e ImageFailedEvent
			if err := json.Unmarshal(msg.Value, &e); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}
			return handler(ctx, e)
		}, nil)
	}()
}

// consume — общий цикл чтения с retry и явным коммитом.
// onDeadLetter вызывается после исчерпания retry; если nil — сообщение просто пропускается.
func (b *EventBus) consume(
	ctx context.Context,
	r *kafka.Reader,
	maxRetries int,
	handler func(context.Context, kafka.Message) error,
	onDeadLetter func(context.Context, kafka.Message, error, int),
) {
	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				b.log.Info("consumer stoping", "topic", r.Config().Topic)
				return
			}
			b.log.Info("fetching failed", "topic", r.Config().Topic, "err", err)
			return
		}

		var lastErr error
		for attempt := 1; attempt <= maxRetries; attempt++ {
			if lastErr = handler(ctx, msg); lastErr == nil {
				break
			}

			b.log.Error("consumer handler failed",
				"topic", r.Config().Topic,
				"key", string(msg.Key),
				"attempt", attempt,
				"err", lastErr)

			if attempt < maxRetries {
				pause := time.Duration(100 * (1 << attempt) * time.Millisecond)
				select {
				case <-ctx.Done():
					return
				case <-time.After(pause):
				}
			}
		}

		if lastErr != nil || onDeadLetter != nil {
			onDeadLetter(ctx, msg, lastErr, maxRetries)
		}

		if err := r.CommitMessages(ctx, msg); err != nil {
			b.log.Error("commit messages failed", "topic", r.Config().Topic, "err", err)
		}
	}
}
