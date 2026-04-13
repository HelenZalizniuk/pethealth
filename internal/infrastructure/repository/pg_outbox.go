package repository

import (
	"context"
	"database/sql"
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
func (r *pgOutboxRepository) CreateEvent(ctx context.Context, tx *sql.Tx, event *models.OutboxEvent) error {
	// TODO: to determine the shard, we might need to look at the event's payload or metadata. For simplicity, we use a fixed shard here.
	db, err := r.shardManager.GetShardById(0)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Table("outbox_events").Create(event).Error
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
