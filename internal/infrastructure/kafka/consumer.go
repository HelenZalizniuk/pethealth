package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MetricConsumer struct {
	reader   *kafka.Reader
	producer *MetricProducer
	logger   *zap.Logger
	dlqTopic string
}

func NewMetricConsumer(brokers []string, topic, dlqTopic, groupID string, producer *MetricProducer, logger *zap.Logger) *MetricConsumer {
	return &MetricConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		producer: producer,
		dlqTopic: dlqTopic,
		logger:   logger,
	}
}

func (c *MetricConsumer) Start(ctx context.Context, handler func(ctx context.Context, msg kafka.Message) error) {
	c.logger.Info("Starting Metric consumer",
		zap.String("topic", c.reader.Config().Topic))

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				c.logger.Info("Consumer stopped: context canceled")
				return
			}

			if err.Error() == "EOF" || err.Error() == "fetching message: EOF" {
				c.logger.Info("Consumer connection closed")
				return
			}

			c.logger.Error("Error reading message from Kafka", zap.Error(err))
			continue
		}

		c.logger.Info("Received message from Kafka",
			zap.String("key", string(m.Key)),
			zap.String("payload", string(m.Value)),
			zap.Int64("offset", m.Offset),
		)

		c.processWithRetry(ctx, m, handler)
	}
}

func (c *MetricConsumer) processWithRetry(ctx context.Context, msg kafka.Message, handler func(ctx context.Context, msg kafka.Message) error) {
	const maxAttempts = 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := handler(ctx, msg); err == nil {
			return
		} else {
			lastErr = err
			c.logger.Warn("Processing attempt failed",
				zap.Int("attempt", attempt),
				zap.String("key", string(msg.Key)),
				zap.Error(err),
			)
		}

		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	c.sendToDLQ(ctx, msg, lastErr)
}

func (c *MetricConsumer) sendToDLQ(ctx context.Context, msg kafka.Message, originalErr error) {
	c.logger.Error("All attempts failed. Sending message to DLQ",
		zap.String("key", string(msg.Key)),
		zap.Error(originalErr),
	)

	err := c.producer.SendToTopic(ctx, c.dlqTopic, msg.Key, msg.Value)
	if err != nil {

		c.logger.Error("CRITICAL: Failed to send message to DLQ",
			zap.String("dlq_topic", c.dlqTopic),
			zap.Error(err),
		)
	}
}

func (c *MetricConsumer) Close() error {
	return c.reader.Close()
}
