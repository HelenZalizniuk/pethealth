package repository

import (
	"context"
	"database/sql"
	"fmt"
	"pethealth/internal/domain/models"
	"pethealth/internal/infrastructure/db"
	"pethealth/internal/infrastructure/metrics"

	"gorm.io/gorm"
)

type PgOutboxRepository struct {
	shardManager *db.ShardManager
}

func NewPGOutboxRepository(sm *db.ShardManager) *PgOutboxRepository {
	return &PgOutboxRepository{shardManager: sm}
}

// save event in the same transaction as the metric
func (r *PgOutboxRepository) CreateEvent(ctx context.Context, tx *sql.Tx, event *models.OutboxEvent, shardingKey uint64) error {

	dbConn, err := r.shardManager.GetShardById(shardingKey)
	if err != nil {
		return fmt.Errorf("failed to route outbox event to shard: %w", err)
	}

	return dbConn.WithContext(ctx).Table("outbox_events").Create(event).Error
}

// get events with pending locking them from other workers
func (r *PgOutboxRepository) FetchAndLockPending(ctx context.Context, shardID int, limit int) ([]models.OutboxEvent, error) {
	var events []models.OutboxEvent
	shardName := r.shardManager.GetShardName(uint64(shardID))

	err := metrics.ObserveDBQuery(shardName, "outbox_fetch_lock", func() error {
		dbConn, err := r.shardManager.GetShardById(uint64(shardID))
		if err != nil {
			return err
		}

		// using transaction with FOR UPDATE
		return dbConn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// SKIP LOCKED allows multiple workers to fetch from the same shard without blocking each other
			rawSQL := `
			SELECT * FROM outbox_events 
			WHERE status = 'pending' 
			LIMIT ? 
			FOR UPDATE SKIP LOCKED`

			return tx.Raw(rawSQL, limit).Scan(&events).Error
		})
	})

	return events, err
}

func (r *PgOutboxRepository) MarkAsPublished(ctx context.Context, id string, petID uint64) error {
	dbConn, err := r.shardManager.GetShardById(petID)
	if err != nil {
		return err
	}

	return dbConn.WithContext(ctx).
		Table("outbox_events").
		Where("id = ?", id).
		Update("status", "published").Error
}
