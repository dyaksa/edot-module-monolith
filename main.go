package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyaksa/warehouse/api/middleware"
	"github.com/dyaksa/warehouse/api/route"
	"github.com/dyaksa/warehouse/api/worker"
	"github.com/dyaksa/warehouse/bootstrap"
	_ "github.com/dyaksa/warehouse/docs" // Swagger docs
	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/dyaksa/warehouse/repository"
	"github.com/dyaksa/warehouse/usecase"
	"github.com/gin-contrib/cors"
)

// @title Warehouse Management API
// @version 1.0
// @description This is a comprehensive warehouse management system API for handling products, orders, warehouses, and transfers.
// @termsOfService http://swagger.io/terms/

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @schemes http https
// @produce json
// @consumes json

func main() {
	ctx := context.Background()
	app := bootstrap.App(ctx)

	defer app.CloseConnection()

	env := app.Env
	router := app.App
	l := app.Log
	db := app.Postgres
	crypto := app.Crypto

	router.Use(cors.Default())
	router.Use(middleware.RateLimitMiddleware(time.Second, 100, "api"))

	timeout := time.Duration(env.ContextTimeout) * time.Second

	reservationRepo := repository.NewReservationRepository(db)
	productStockRepo := repository.NewProductStockRepository(db)
	movementRepo := repository.NewMovementRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// Initialize stock release usecase
	stockReleaseUsecase := usecase.NewStockReleaseUsecase(
		db.Database(),
		reservationRepo,
		productStockRepo,
		movementRepo,
		orderRepo,
	)

	workerConfig := worker.StockReleaseWorkerConfig{
		BatchSize: 50,
		Interval:  30 * time.Second,
	}

	stockReleaseWorker := worker.NewStockReleaseWorker(stockReleaseUsecase, workerConfig)

	// Create context for worker with cancellation
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	// Start the stock release worker in a goroutine
	go func() {
		l.Info("Starting stock release worker...")
		stockReleaseWorker.Start(workerCtx)
	}()

	route.Setup(env, timeout, db, l, crypto, router)

	route.NewStockReleaseRoute(env, timeout, db, l, crypto, router, stockReleaseWorker)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", env.Port),
		Handler: app.App.Handler(),
	}

	go func() {
		l.Info(fmt.Sprintf("%s is running on port %s", env.AppName, env.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("failed to start server", log.Error("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("shutting down server")

	l.Info("stopping stock release worker")
	workerCancel() // Cancel the worker context
	stockReleaseWorker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		l.Fatal("server forced to shutdown", log.Error("error", err))
	}

	l.Info("server stopped")

}
