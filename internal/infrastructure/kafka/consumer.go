package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MetricConsumer struct {
	reader  *kafka.Reader
	logger  *zap.Logger
	brokers []string
}

func NewMetricConsumer(brokers []string, topic string, groupID string, logger *zap.Logger) *MetricConsumer {
	return &MetricConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		logger:  logger,
		brokers: brokers,
	}
}

func (c *MetricConsumer) Start(ctx context.Context) {
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
		//TODO: process the message
		c.logger.Info("Received message from Kafka",
			zap.String("key", string(m.Key)),
			zap.String("payload", string(m.Value)),
			zap.Int64("offset", m.Offset),
		)
	}
}

func (c *MetricConsumer) Close() error {
	return c.reader.Close()
}
