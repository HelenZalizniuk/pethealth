package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"pethealth/internal/config"
	"pethealth/internal/infrastructure/db"
	"pethealth/internal/infrastructure/transport"
	"pethealth/internal/infrastructure/transport/middleware"
	"pethealth/internal/infrastructure/worker"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	Cfg           *config.Config
	MetricHandler *transport.MetricHandler
	Logger        *zap.Logger
	sm            *db.ShardManager
	relayPool     *worker.WorkerPool
}

// root initializer for the application
func NewApp(cfg *config.Config, handler *transport.MetricHandler, l *zap.Logger, sm *db.ShardManager, relayPool *worker.WorkerPool) *App {

	return &App{
		Cfg:           cfg,
		MetricHandler: handler,
		Logger:        l,
		sm:            sm,
		relayPool:     relayPool,
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

	// 1. Setup Relay Workers
	// We create a separate context to manage the lifecycle of background workers
	relayCtx, cancelRelay := context.WithCancel(context.Background())
	defer cancelRelay()

	a.Logger.Info("Starting Outbox Relay workers...")
	go a.relayPool.Start(relayCtx)

	// 2. Setup HTTP Server (GIN)
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

	// Stop Relay workers first to ensure all polled events are processed
	a.Logger.Info("Stopping Outbox Relay workers...")
	cancelRelay()      // Send cancellation signal to workers
	a.relayPool.Stop() // Wait for workers to finish their current batch

	// Then shutdown the HTTP server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.Logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	a.Logger.Info("Server exited gracefully")
	return nil
}
