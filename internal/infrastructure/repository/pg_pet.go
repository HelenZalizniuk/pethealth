package repository

import (
	"context"
	"pethealth/internal/domain/models"
	"pethealth/internal/infrastructure/db"
)

type pgPetRepository struct {
	shardManager *db.ShardManager
}

func NewPGPetRepository(sm *db.ShardManager) *pgPetRepository {
	return &pgPetRepository{shardManager: sm}
}

func (r *pgPetRepository) Create(ctx context.Context, pet *models.Pet) error {
	// sharding by OwnerID
	db, err := r.shardManager.GetShardById(pet.OwnerID)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Create(pet).Error
}

func (r *pgPetRepository) GetByID(ctx context.Context, id uint64) (*models.Pet, error) {
	// Usually pet ID has ID-sharding
	db, err := r.shardManager.GetShardById(id)
	if err != nil {
		return nil, err
	}

	var pet models.Pet
	err = db.WithContext(ctx).First(&pet, id).Error
	return &pet, err
}
