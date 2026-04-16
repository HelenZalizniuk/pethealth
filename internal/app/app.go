package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pethealth/internal/config"
	"pethealth/internal/infrastructure/transport"
	"pethealth/internal/infrastructure/transport/middleware"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	Cfg           *config.Config
	MetricHandler *transport.MetricHandler
	Logger        *zap.Logger
}

// root initializer for the application
func NewApp(cfg *config.Config, handler *transport.MetricHandler, l *zap.Logger) *App {

	return &App{
		Cfg:           cfg,
		MetricHandler: handler,
		Logger:        l,
	}
}

// Run starts the application components
func (a *App) Run() error {
	log.Printf("PetHealth Service is starting on port %s...", a.Cfg.ServerPort)

	// HTTP Server GIN
	r := gin.Default()

	// Global Middleware
	r.Use(middleware.RequestIDMiddleware())

	// Health-check for future Kubernetes integration
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// API-metrics group
	api := r.Group("/api/v1")
	{

		api.POST("/metrics", a.MetricHandler.ReceiveMetric)
	}
	// for graceful shutdown
	srv := &http.Server{
		Addr:    ":" + a.Cfg.ServerPort,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatal("Server startup failed", zap.Error(err))
		}
	}()

	a.Logger.Info("Application started", zap.String("port", a.Cfg.ServerPort))

	// wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // lock until signal is received

	a.Logger.Info("Shutting down...")
	// graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.Logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	a.Logger.Info("Server exited gracefully")
	return nil
}
