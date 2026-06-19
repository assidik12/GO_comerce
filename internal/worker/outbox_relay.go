package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/event"
)

type OutboxRelay interface {
	Start(ctx context.Context)
}

type outboxRelay struct {
	repo     domain.OutboxRepository
	producer event.Producer
	logger   *slog.Logger
	interval time.Duration
}

func NewOutboxRelay(repo domain.OutboxRepository, producer event.Producer, logger *slog.Logger) OutboxRelay {
	return &outboxRelay{
		repo:     repo,
		producer: producer,
		logger:   logger,
		interval: 3 * time.Second, // Poll every 3 seconds
	}
}

func (r *outboxRelay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.logger.Info("Outbox relay worker started")

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Outbox relay worker stopped")
			return
		case <-ticker.C:
			r.processOutbox(ctx)
		}
	}
}

func (r *outboxRelay) processOutbox(ctx context.Context) {
	// 1. Fetch pending events
	events, err := r.repo.FindPending(ctx, 50)
	if err != nil {
		r.logger.Error("failed to fetch pending outbox events", "error", err)
		return
	}

	if len(events) == 0 {
		return
	}

	// 2. Publish to Kafka
	var processedIDs []string
	for _, evt := range events {
		// Because payload is already JSON, we can publish it as raw bytes or unmarshal to the struct.
		// Our producer.Publish accepts an interface{}, so we should probably unmarshal or just modify the producer to take bytes.
		// Wait, the current producer takes interface{} and marshals it. If we pass raw JSON bytes, it might double marshal.
		// We can unmarshal to a map[string]interface{} or specifically `OrderCreatedEvent`.
		// Since we know the topic, we can just pass it as is if the producer is modified, or we can just unmarshal it.
		raw := json.RawMessage(evt.Payload)

		pubCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := r.producer.Publish(pubCtx, evt.Topic, raw)
		cancel()

		if err != nil {
			r.logger.Error("failed to publish outbox event", "id", evt.ID, "topic", evt.Topic, "error", err)
			continue
		}
		processedIDs = append(processedIDs, evt.ID)
	}

	// 3. Mark as processed
	if len(processedIDs) > 0 {
		if err := r.repo.MarkAsProcessed(ctx, processedIDs); err != nil {
			r.logger.Error("failed to mark outbox events as processed", "error", err)
		} else {
			r.logger.Info("processed outbox events", "count", len(processedIDs))
		}
	}
}
