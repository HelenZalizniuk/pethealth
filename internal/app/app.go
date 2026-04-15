package app

import (
	"fmt"
	"log"
	"pethealth/internal/config"
	"pethealth/internal/infrastructure/db"
	"pethealth/internal/infrastructure/service"
	"pethealth/internal/usecase"

	infraRepo "pethealth/internal/infrastructure/repository"
)

type App struct {
	Cfg           *config.Config
	MetricUseCase *usecase.MetricUseCase
}

// root initializer for the application
func NewApp(cfg *config.Config) (*App, error) {
	// 1. Init Infrastructure (Database Shards)
	shardManager, err := db.NewShardManager(cfg.Shards)
	if err != nil {
		return nil, fmt.Errorf("failed to init shard manager: %w", err)
	}

	// 2. Init Repositories
	metricRepo := infraRepo.NewPGHealthMetricRepository(shardManager)
	outboxRepo := infraRepo.NewPGOutboxRepository(shardManager)

	// 3. Init Services
	thresholdService := service.NewStaticThresholdService(150.0)

	// 4. Init UseCases
	metricUseCase := usecase.NewMetricUseCase(metricRepo, outboxRepo, thresholdService)

	return &App{
		Cfg:           cfg,
		MetricUseCase: metricUseCase,
	}, nil
}

// Run starts the application components
func (a *App) Run() error {
	log.Println("PetHealth Service is starting...")

	// TODO: HTTP server initialization.

	log.Println("System is ready and waiting for metrics.")
	return nil
}
