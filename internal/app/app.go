package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"pethealth/internal/config"
	"pethealth/internal/infrastructure/db"
	infrakafka "pethealth/internal/infrastructure/kafka"
	"pethealth/internal/infrastructure/transport"
	"pethealth/internal/infrastructure/transport/middleware"
	"pethealth/internal/infrastructure/worker"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type App struct {
	Cfg           *config.Config
	MetricHandler *transport.MetricHandler
	Logger        *zap.Logger
	sm            *db.ShardManager
	relayPool     *worker.WorkerPool
	consumer      *infrakafka.MetricConsumer
}

// root initializer for the application
func NewApp(cfg *config.Config,
	handler *transport.MetricHandler,
	logger *zap.Logger,
	sm *db.ShardManager,
	relayPool *worker.WorkerPool,
	producer *infrakafka.MetricProducer,
) *App {

	consumer := infrakafka.NewMetricConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
		cfg.KafkaDLQTopic,
		cfg.KafkaConsumerGroup,
		producer,
		logger)
	return &App{
		Cfg:           cfg,
		MetricHandler: handler,
		Logger:        logger,
		sm:            sm,
		relayPool:     relayPool,
		consumer:      consumer,
	}
}

func (a *App) checkDependencies() error {

	if a.sm == nil {
		return fmt.Errorf("shard manager is not initialized")
	}

	return a.sm.Ping()
}

// Run starts the application components
func (a *App) Run() error {
	a.Logger.Info("PetHealth Service is starting...", zap.String("port", a.Cfg.ServerPort))

	// Context for background processes
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	// Start Relay
	a.Logger.Info("Starting Outbox Relay workers...")
	go a.relayPool.Start(workerCtx)

	// Start Kafka Consumer
	a.Logger.Info("Starting Kafka consumer...")
	go a.consumer.Start(workerCtx, func(ctx context.Context, msg kafka.Message) error {
		return a.MetricHandler.ProcessMetric(ctx, msg.Value)
	})

	// Setup HTTP Server (GIN)
	r := gin.Default()
	// Global Middleware
	r.Use(middleware.RequestIDMiddleware())

	// Liveness probe
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Readiness probe
	r.GET("/ready", func(c *gin.Context) {
		if err := a.checkDependencies(); err != nil {
			a.Logger.Warn("Readiness check failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "NOT_READY",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "READY"})
	})

	// Prometheus Metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API-metrics group
	api := r.Group("/api/v1")
	{
		api.POST("/metrics", a.MetricHandler.ReceiveMetric)
	}

	// Server configuration
	srv := &http.Server{
		Addr:    ":" + a.Cfg.ServerPort,
		Handler: r,
	}

	// Start HTTP server in a separate goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatal("Server startup failed", zap.Error(err))
		}
	}()

	a.Logger.Info("Application started", zap.String("port", a.Cfg.ServerPort))

	// 3. Graceful Shutdown Management
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until signal is received

	a.Logger.Info("Shutting down PetHealth Service...")

	a.Logger.Info("Stopping background workers (Relay and Consumer)...")
	cancelWorkers()

	// close physical connections
	if err := a.consumer.Close(); err != nil {
		a.Logger.Error("Failed to close Kafka consumer", zap.Error(err))
	}

	// wait for Relay to finish processing current batch
	a.relayPool.Stop()

	// Shutdown the HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		a.Logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	a.Logger.Info("Server exited gracefully")
	return nil
}
