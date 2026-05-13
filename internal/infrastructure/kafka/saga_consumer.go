package kafka

import (
	"context"
	"encoding/json"
	"log"
	"pethealth/internal/config"

	"github.com/segmentio/kafka-go"
)

type SagaResponse struct {
	PetID   string `json:"pet_id"`
	Success bool   `json:"success"`
}

type PetStatusUpdater interface {
	UpdateStatus(ctx context.Context, petID string, status string) error
}

type TaskSubmitter interface {
	Submit(task func(ctx context.Context))
}

type SagaResponseConsumer struct {
	reader     *kafka.Reader
	repo       PetStatusUpdater
	workerPool TaskSubmitter
}

func NewSagaResponseConsumer(cfg *config.Config, repo PetStatusUpdater, pool TaskSubmitter) *SagaResponseConsumer {
	readerCfg := kafka.ReaderConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaSagaTopic,
		GroupID: cfg.KafkaSagaConsumerGroup,
	}

	return &SagaResponseConsumer{
		reader:     kafka.NewReader(readerCfg),
		repo:       repo,
		workerPool: pool,
	}
}

func (c *SagaResponseConsumer) Start(ctx context.Context) {
	log.Println("Saga Response Consumer started and listening to Kafka...")

	for {
		// Reading messages (blocking call)
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading from Kafka: %v", err)
			continue
		}

		// handling message in a separate goroutine to avoid blocking the consumer loop
		c.workerPool.Submit(func(ctx context.Context) {
			c.processMessage(ctx, msg.Value)
		})
	}
}

func (c *SagaResponseConsumer) processMessage(ctx context.Context, data []byte) {
	var resp SagaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	finalStatus := "ACTIVE"
	if !resp.Success {
		finalStatus = "FAILED" // Compensation transaction (rollback)
	}

	err := c.repo.UpdateStatus(ctx, resp.PetID, finalStatus)
	if err != nil {
		log.Printf("Error updating status for PetID %s: %v", resp.PetID, err)
		return
	}

	log.Printf("Status for pet %s successfully updated to %s", resp.PetID, finalStatus)
}

func (c *SagaResponseConsumer) Close() error {
	log.Println("Закрытие Saga Response Consumer...")
	return c.reader.Close()
}
