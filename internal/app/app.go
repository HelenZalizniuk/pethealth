package app

import (
	"log"
	"net/http"
	"pethealth/internal/config"
	"pethealth/internal/infrastructure/transport"
	"pethealth/internal/usecase"

	infraRepo "pethealth/internal/infrastructure/repository"

	"github.com/gin-gonic/gin"
)

type App struct {
	Cfg           *config.Config
	MetricHandler *transport.MetricHandler
}

// root initializer for the application
func NewApp(cfg *config.Config) (*App, error) {
	// 1. Init Infrastructure (Database Shards)
	// shardManager, err := db.NewShardManager(cfg.Shards)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to init shard manager: %w", err)
	// }

	// 2. Init Repositories
	// metricRepo := infraRepo.NewPGHealthMetricRepository(shardManager)
	// outboxRepo := infraRepo.NewPGOutboxRepository(shardManager)
	mRepo := &infraRepo.MockMetricRepository{}
	oRepo := &infraRepo.MockOutboxRepository{}
	tService := &infraRepo.MockThresholdService{}

	// 3. Init Services
	// thresholdService := service.NewStaticThresholdService(150.0)

	// 4. Init UseCases
	// metricUseCase := usecase.NewMetricUseCase(metricRepo, outboxRepo, thresholdService)
	metricUseCase := usecase.NewMetricUseCase(mRepo, oRepo, tService)

	// 5. Init Handlers
	metricHandler := transport.NewMetricHandler(metricUseCase)

	return &App{
		Cfg:           cfg,
		MetricHandler: metricHandler,
	}, nil
}

// Run starts the application components
func (a *App) Run() error {
	log.Printf("PetHealth Service is starting on port %s...", a.Cfg.ServerPort)

	// HTTP Server setup
	r := gin.Default()

	// API-metrics group
	api := r.Group("/api/v1")
	{
		api.POST("/metrics", a.MetricHandler.ReceiveMetric)
	}

	// Health-check for future Kubernetes integration
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	return r.Run(":" + a.Cfg.ServerPort)
}
