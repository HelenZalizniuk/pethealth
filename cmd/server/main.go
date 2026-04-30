package main

import (
	"context"
	"log"
	"pethealth/internal/app"
	"pethealth/internal/config"
	"pethealth/internal/infrastructure/db"
	"pethealth/internal/infrastructure/kafka"
	"pethealth/internal/infrastructure/logger"
	"pethealth/internal/infrastructure/repository"
	"pethealth/internal/infrastructure/service"
	"pethealth/internal/infrastructure/transport"
	"pethealth/internal/infrastructure/validator"
	"pethealth/internal/infrastructure/worker"
	"pethealth/internal/usecase"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// load .env file (if exists) - useful for local development and testing
	// vars in prod can be set in K8s)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// load configuration
	cfg := config.Load()
	logger := logger.NewLogger()
	defer logger.Sync()

	v := validator.NewCustomValidator()

	// Data & business layer setup
	sm, err := db.NewShardManager(cfg.Shards)
	if err != nil {
		logger.Fatal("Failed to initialize ShardManager", zap.Error(err))
	}

	metricRepo := repository.NewPGHealthMetricRepository(sm)
	outboxRepo := repository.NewPGOutboxRepository(sm)
	thresholds := service.NewStaticThresholdService(140.0)

	uc := usecase.NewMetricUseCase(metricRepo, outboxRepo, thresholds)

	handler := transport.NewMetricHandler(uc, v, logger)

	producer := kafka.NewMetricProducer(cfg.KafkaBrokers, cfg.KafkaTopic, logger)
	defer producer.Close()

	// Initialize required Kafka topics before starting the application
	ctx := context.Background()
	if err := producer.EnsureTopicExists(ctx, cfg.KafkaTopic); err != nil {
		logger.Fatal("Failed to ensure Kafka topic exists",
			zap.String("topic", cfg.KafkaTopic),
			zap.Error(err),
		)
	}

	relayProcessor := worker.NewOutboxProcessor(outboxRepo, producer, logger)
	relayPool := worker.NewWorkerPool(2, relayProcessor, logger)

	// initialize
	application := app.NewApp(cfg, handler, logger, sm, relayPool, producer)

	// run the application (starts HTTP server, etc.)
	if err := application.Run(); err != nil {
		log.Fatalf("Application stopped with error: %v", err)
	}
}
