package main

import (
	"log"
	"pethealth/internal/app"
	"pethealth/internal/config"
	"pethealth/internal/infrastructure/db"
	"pethealth/internal/infrastructure/logger"
	"pethealth/internal/infrastructure/repository"
	"pethealth/internal/infrastructure/service"
	"pethealth/internal/infrastructure/transport"
	"pethealth/internal/infrastructure/validator"
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
	l := logger.NewLogger()
	defer l.Sync()

	v := validator.NewCustomValidator()

	// Data & business layer setup
	sm, err := db.NewShardManager(cfg.Shards)
	if err != nil {
		l.Fatal("Failed to initialize ShardManager", zap.Error(err))
	}

	metricRepo := repository.NewPGHealthMetricRepository(sm)
	outboxRepo := repository.NewPGOutboxRepository(sm)
	thresholds := service.NewStaticThresholdService(140.0)

	uc := usecase.NewMetricUseCase(metricRepo, outboxRepo, thresholds)

	handler := transport.NewMetricHandler(uc, v, l)

	// initialize
	application := app.NewApp(cfg, handler, l)

	// run the application (starts HTTP server, etc.)
	if err := application.Run(); err != nil {
		log.Fatalf("Application stopped with error: %v", err)
	}
}
