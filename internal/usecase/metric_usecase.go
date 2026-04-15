package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pethealth/internal/domain/models"
	"pethealth/internal/domain/repository"

	"github.com/google/uuid"
)

const (
	DefaultThresholdHeartRate = 150.0
	Alert                     = "Threshold Exceeded"
	TopicHealthAlerts         = "health_alerts"
	StatusPending             = "pending"
)

type MetricUseCase struct {
	metricRepo repository.HealthMetricRepository
	outboxRepo repository.OutboxRepository
	thresholds repository.ThresholdProvider
}

func NewMetricUseCase(mr repository.HealthMetricRepository,
	or repository.OutboxRepository,
	tp repository.ThresholdProvider) *MetricUseCase {
	return &MetricUseCase{
		metricRepo: mr,
		outboxRepo: or,
		thresholds: tp,
	}
}

func (u *MetricUseCase) ProcessMetric(ctx context.Context, metric *models.HealthMetric) error {

	if err := u.metricRepo.Store(ctx, metric); err != nil {
		return fmt.Errorf("failed to store metric: %w", err)
	}

	// getting dynamic threshold for the metric type and pet
	limit, err := u.thresholds.GetThreshold(ctx, metric.PetID, metric.Type)
	if err != nil {
		// default
		limit = DefaultThresholdHeartRate
	}

	if metric.Value > limit {
		payload, _ := json.Marshal(map[string]interface{}{
			"pet_id":      metric.PetID,
			"value":       metric.Value,
			"threshold":   limit,
			"metric_type": metric.Type,
			"alert":       Alert,
		})

		event := &models.OutboxEvent{
			ID:        uuid.New().String(),
			PetID:     metric.PetID,
			Payload:   payload,
			Topic:     TopicHealthAlerts,
			Status:    StatusPending,
			CreatedAt: time.Now(),
		}

		// metric.PetID as shardingKey
		return u.outboxRepo.CreateEvent(ctx, nil, event, metric.PetID)
	}

	return nil
}
