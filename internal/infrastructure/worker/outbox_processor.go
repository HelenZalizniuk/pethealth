package worker

import (
	"context"
	"pethealth/internal/infrastructure/kafka"
	"pethealth/internal/infrastructure/repository"

	"go.uber.org/zap"
)

type OutboxProcessor struct {
	repo     *repository.PgOutboxRepository
	producer *kafka.MetricProducer
	logger   *zap.Logger
}

func NewOutboxProcessor(repo *repository.PgOutboxRepository, producer *kafka.MetricProducer, logger *zap.Logger) *OutboxProcessor {
	return &OutboxProcessor{
		repo:     repo,
		producer: producer,
		logger:   logger,
	}
}

func (p *OutboxProcessor) ProcessNextBatch(ctx context.Context, workerID int) {
	shardID := workerID % 2

	events, err := p.repo.FetchAndLockPending(ctx, shardID, 10)
	if err != nil {
		p.logger.Error("Failed to fetch pending events",
			zap.Int("shard", shardID),
			zap.Error(err),
		)
		return
	}

	for _, event := range events {

		// send to Kafka
		if err := p.producer.SendEvent(ctx, &event); err != nil {
			p.logger.Error("Failed to send to Kafka",
				zap.String("event_id", event.ID),
				zap.Error(err))
			continue
		}

		// mark event as published
		if err := p.repo.MarkAsPublished(ctx, event.ID, event.PetID); err != nil {
			p.logger.Error("Failed to mark event as published",
				zap.String("event_id", event.ID),
				zap.Error(err))
		}
	}
}
