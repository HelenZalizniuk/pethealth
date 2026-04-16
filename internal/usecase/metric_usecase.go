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

func NewMetricUseCase(
	mr repository.HealthMetricRepository,
	or repository.OutboxRepository,
	tp repository.ThresholdProvider) *MetricUseCase {
	return &MetricUseCase{
		metricRepo: mr,
		outboxRepo: or,
		thresholds: tp,
	}
}

func (u *MetricUseCase) ProcessMetric(ctx context.Context, metric *models.HealthMetric) error {

	rid, _ := ctx.Value("RequestID").(string)

	if err := u.metricRepo.Store(ctx, metric); err != nil {
		return fmt.Errorf("[RID: %s] failed to store metric: %w", rid, err)
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
			"external_id": metric.ExternalID,
			"alert":       Alert,
		})

		event := &models.OutboxEvent{
			ID:        uuid.New().String(),
			PetID:     metric.PetID,
			Type:      "alert_triggered",
			Payload:   payload,
			Topic:     TopicHealthAlerts,
			Status:    StatusPending,
			CreatedAt: time.Now(),
		}

		// metric.PetID as shardingKey
		if err := u.outboxRepo.CreateEvent(ctx, nil, event, metric.PetID); err != nil {
			return fmt.Errorf("[RID: %s] failed to create outbox event: %w", rid, err)
		}
	}

	return nil
}
