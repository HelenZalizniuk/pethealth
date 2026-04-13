package repository

import (
	"context"
	"pethealth/internal/domain/models"
	"pethealth/internal/infrastructure/db"
)

type pgHealthMetricRepository struct {
	shardManager *db.ShardManager
}

func NewPGHealthMetricRepository(sm *db.ShardManager) *pgHealthMetricRepository {
	return &pgHealthMetricRepository{shardManager: sm}
}

func (r *pgHealthMetricRepository) Store(ctx context.Context, metric *models.HealthMetric) error {
	// sharding by PetID, to ensure all metrics for a pet go to the same shard
	db, err := r.shardManager.GetShardById(metric.PetID)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Create(metric).Error
}

func (r *pgHealthMetricRepository) GetByPetID(ctx context.Context, petID uint64, limit int) ([]models.HealthMetric, error) {
	db, err := r.shardManager.GetShardById(petID)
	if err != nil {
		return nil, err
	}

	var metrics []models.HealthMetric
	// using replica for reads via dbresolver
	err = db.WithContext(ctx).
		Where("pet_id = ?", petID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&metrics).Error

	return metrics, err
}
