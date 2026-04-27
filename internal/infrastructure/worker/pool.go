package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type WorkerPool struct {
	workerCount int
	processor   *OutboxProcessor
	logger      *zap.Logger
	wg          sync.WaitGroup
}

func NewWorkerPool(count int, processor *OutboxProcessor, logger *zap.Logger) *WorkerPool {
	return &WorkerPool{
		workerCount: count,
		processor:   processor,
		logger:      logger,
	}
}

func (wp *WorkerPool) oldStart(ctx context.Context, task func(ctx context.Context)) {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go func(workerID int) {
			defer wp.wg.Done()
			wp.logger.Info("Worker started", zap.Int("worker_id", workerID))
			task(ctx)
			wp.logger.Info("Worker stopped", zap.Int("worker_id", workerID))
		}(i)
	}
}

// starting workers in background mode
func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go func(id int) {
			defer wp.wg.Done()

			// adding shift of 100ms per worker ID to avoid them hitting the database at the same time
			interval := 500 * time.Millisecond
			ticker := time.NewTicker(interval + time.Duration(id*100)*time.Millisecond)
			defer ticker.Stop()

			wp.logger.Info("Outbox Relay worker started", zap.Int("id", id))

			for {
				select {
				case <-ctx.Done():
					wp.logger.Info("Outbox Relay worker received shutdown signal", zap.Int("id", id))
					return
				case <-ticker.C:
					wp.processor.ProcessNextBatch(ctx, id)
				}
			}
		}(i)
	}
}

func (wp *WorkerPool) Stop() {
	wp.logger.Info("Waiting for Outbox Relay workers to finish...")
	wp.wg.Wait()
	wp.logger.Info("All Outbox Relay workers stopped successfully")
}
