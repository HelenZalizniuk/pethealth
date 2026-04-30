package kafka

import (
	"context"
	"fmt"
	"pethealth/internal/domain/models"
	"strconv"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MetricProducer struct {
	writer  *kafka.Writer
	logger  *zap.Logger
	brokers []string
}

func NewMetricProducer(brokers []string, topic string, l *zap.Logger) *MetricProducer {
	return &MetricProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			Async:        false,
		},
		logger:  l,
		brokers: brokers,
	}
}

func (p *MetricProducer) SendEvent(ctx context.Context, event *models.OutboxEvent) error {

	key := strconv.FormatUint(event.PetID, 10)

	err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: event.Topic,
		Key:   []byte(key),
		Value: event.Payload,
	})

	if err != nil {
		p.logger.Error("Failed to send message to Kafka", zap.Error(err), zap.Uint64("pet_id", event.PetID))
		return err
	}

	return nil
}

func (p *MetricProducer) Close() error {
	return p.writer.Close()
}

func (p *MetricProducer) EnsureTopicExists(ctx context.Context, topic string) error {

	if len(p.brokers) == 0 {
		return fmt.Errorf("no kafka brokers configured")
	}

	conn, err := kafka.DialContext(ctx, "tcp", p.brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(topic)
	if err == nil && len(partitions) > 0 {
		p.logger.Debug("Topic already exists", zap.String("topic", topic))
		return nil
	}

	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	if err := conn.CreateTopics(topicConfig); err != nil {
		return fmt.Errorf("failed to create topic %s: %w", topic, err)
	}

	p.logger.Info("Successfully created kafka topic", zap.String("topic", topic))
	return nil
}

func (p *MetricProducer) SendToTopic(ctx context.Context, topic string, key []byte, value []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
	if err != nil {
		p.logger.Error("Failed to send message to DLQ",
			zap.String("topic", topic),
			zap.Error(err))
		return err
	}
	return nil
}
