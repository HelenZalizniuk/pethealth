package repository

import (
	"context"
	"fmt"
	"pethealth/internal/domain/models"
	"pethealth/internal/infrastructure/db"
	"strconv"

	"gorm.io/gorm"
)

type PgPetRepository struct {
	shardManager *db.ShardManager
}

func NewPGPetRepository(sm *db.ShardManager) *PgPetRepository {
	return &PgPetRepository{shardManager: sm}
}

func (r *PgPetRepository) Create(ctx context.Context, pet *models.Pet) error {
	// sharding by OwnerID
	db, err := r.shardManager.GetShardById(pet.OwnerID)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Create(pet).Error
}

func (r *PgPetRepository) GetByID(ctx context.Context, id uint64) (*models.Pet, error) {
	// Usually pet ID has ID-sharding
	db, err := r.shardManager.GetShardById(id)
	if err != nil {
		return nil, err
	}

	var pet models.Pet
	err = db.WithContext(ctx).First(&pet, id).Error
	return &pet, err
}

func (r *PgPetRepository) UpdateStatus(ctx context.Context, petIDStr string, status string) error {
	petID, err := strconv.ParseUint(petIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid pet_id format: %w", err)
	}

	db, err := r.shardManager.GetShardById(petID)
	if err != nil {
		return fmt.Errorf("failed to get shard: %w", err)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Pet{}).
			Where("id = ?", petID).
			Update("status", status)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("no pet found with id %d on this shard", petID)
		}
		return nil
	})
}
