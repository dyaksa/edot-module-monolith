package worker

import (
	"context"
	"log"
	"time"

	"github.com/dyaksa/warehouse/usecase"
)

type StockReleaseWorker struct {
	stockReleaseUsecase usecase.StockReleaseUsecase
	batchSize           int
	interval            time.Duration
	stopCh              chan struct{}
}

type StockReleaseWorkerConfig struct {
	BatchSize int           // Number of reservations to process per batch
	Interval  time.Duration // How often to check for expired reservations
}

func NewStockReleaseWorker(
	stockReleaseUsecase usecase.StockReleaseUsecase,
	config StockReleaseWorkerConfig,
) *StockReleaseWorker {
	// Set default values if not provided
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.Interval <= 0 {
		config.Interval = 30 * time.Second // Check every 30 seconds by default
	}

	return &StockReleaseWorker{
		stockReleaseUsecase: stockReleaseUsecase,
		batchSize:           config.BatchSize,
		interval:            config.Interval,
		stopCh:              make(chan struct{}),
	}
}

// Start begins the background worker that periodically processes expired reservations
func (w *StockReleaseWorker) Start(ctx context.Context) {
	log.Printf("Starting stock release worker with batch size %d and interval %v", w.batchSize, w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stock release worker stopped due to context cancellation")
			return
		case <-w.stopCh:
			log.Println("Stock release worker stopped")
			return
		case <-ticker.C:
			w.processExpiredReservations(ctx)
		}
	}
}

// Stop gracefully stops the worker
func (w *StockReleaseWorker) Stop() {
	close(w.stopCh)
}

// StartWithGracefulShutdown starts the worker and handles graceful shutdown
func (w *StockReleaseWorker) StartWithGracefulShutdown(ctx context.Context) {
	go w.Start(ctx)

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Shutting down stock release worker...")
	w.Stop()
}

func (w *StockReleaseWorker) processExpiredReservations(ctx context.Context) {
	start := time.Now()

	err := w.stockReleaseUsecase.ProcessExpiredReservations(ctx, w.batchSize)
	if err != nil {
		log.Printf("Error processing expired reservations: %v", err)
		return
	}

	duration := time.Since(start)
	log.Printf("Processed expired reservations in %v", duration)
}

// ProcessNow immediately processes expired reservations (useful for manual triggers)
func (w *StockReleaseWorker) ProcessNow(ctx context.Context) error {
	log.Println("Manually triggered stock release processing")
	return w.stockReleaseUsecase.ProcessExpiredReservations(ctx, w.batchSize)
}
