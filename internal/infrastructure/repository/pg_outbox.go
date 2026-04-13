package repository

import (
	"context"
	"database/sql"
	"fmt"
	"pethealth/internal/domain/models"
	"pethealth/internal/infrastructure/db"
)

type pgOutboxRepository struct {
	shardManager *db.ShardManager
}

func NewPGOutboxRepository(sm *db.ShardManager) *pgOutboxRepository {
	return &pgOutboxRepository{shardManager: sm}
}

// save event in the same transaction as the metric
func (r *pgOutboxRepository) CreateEvent(ctx context.Context, tx *sql.Tx, event *models.OutboxEvent, shardingKey uint64) error {

	dbConn, err := r.shardManager.GetShardById(shardingKey)
	if err != nil {
		return fmt.Errorf("failed to route outbox event to shard: %w", err)
	}

	return dbConn.WithContext(ctx).Table("outbox_events").Create(event).Error
}

// usually wokrks in background
func (r *pgOutboxRepository) FetchPending(ctx context.Context, limit int) ([]models.OutboxEvent, error) {
	// Тут воркер должен будет пройтись по всем шардам и собрать события.
	// TODO: relay
	return nil, nil
}

func (r *pgOutboxRepository) MarkAsPublished(ctx context.Context, id string) error {
	return nil
}
