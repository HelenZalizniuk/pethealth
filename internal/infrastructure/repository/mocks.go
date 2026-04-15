package repository

import (
	"context"
	"database/sql"
	"log"
	"pethealth/internal/domain/models"
)

// MockMetricRepository implements HealthMetricRepository
type MockMetricRepository struct{}

func (m *MockMetricRepository) Store(ctx context.Context, metric *models.HealthMetric) error {
	log.Printf("[MOCK DB] Store Metric: PetID=%d, Value=%.2f", metric.PetID, metric.Value)
	return nil
}

func (m *MockMetricRepository) GetByPetID(ctx context.Context, petID uint64, limit int) ([]models.HealthMetric, error) {
	return []models.HealthMetric{}, nil
}

// MockOutboxRepository implements OutboxRepository
type MockOutboxRepository struct{}

func (m *MockOutboxRepository) CreateEvent(ctx context.Context, tx *sql.Tx, event *models.OutboxEvent, shardingKey uint64) error {
	log.Printf("[MOCK DB] Create Event: Topic=%s, PetID=%d", event.Topic, shardingKey)
	return nil
}

func (m *MockOutboxRepository) FetchPending(ctx context.Context, limit int) ([]models.OutboxEvent, error) {
	return []models.OutboxEvent{}, nil
}

func (m *MockOutboxRepository) MarkAsPublished(ctx context.Context, id string) error {
	return nil
}

// MockThresholdService implements ThresholdProvider
type MockThresholdService struct{}

func (m *MockThresholdService) GetThreshold(ctx context.Context, petID uint64, metricType string) (float64, error) {
	return 150.0, nil
}
