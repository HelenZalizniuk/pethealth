package repository

import (
	"context"
	"database/sql"
	"pethealth/internal/domain/models"
)

// for working with basic pet data.
type PetRepository interface {
	Create(ctx context.Context, pet *models.Pet) error
	GetByID(ctx context.Context, id uint64) (*models.Pet, error)
}

// contract for high-load metrics recording
type HealthMetricRepository interface {
	// stores metrics (sharding)
	Store(ctx context.Context, metric *models.HealthMetric) error
	// GetByPetID from replica
	GetByPetID(ctx context.Context, petID uint64, limit int) ([]models.HealthMetric, error)
}

type OutboxRepository interface {
	CreateEvent(ctx context.Context, tx *sql.Tx, event *models.OutboxEvent, shardingKey uint64) error
	FetchAndLockPending(ctx context.Context, shardID int, limit int) ([]models.OutboxEvent, error)
	MarkAsPublished(ctx context.Context, id string, petID uint64) error
}

// Transactor — interface for managing transactions within a single shard.
type Transactor interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

type ThresholdProvider interface {
	GetThreshold(ctx context.Context, petID uint64, metricType string) (float64, error)
}
