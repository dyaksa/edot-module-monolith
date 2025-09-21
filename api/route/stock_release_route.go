package route

import (
	"time"

	"github.com/dyaksa/warehouse/api/controller"
	"github.com/dyaksa/warehouse/api/middleware"
	"github.com/dyaksa/warehouse/api/worker"
	"github.com/dyaksa/warehouse/bootstrap"
	"github.com/dyaksa/warehouse/infrastructure/crypto"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/gin-gonic/gin"
)

func NewStockReleaseRoute(
	env *bootstrap.Env,
	timeout time.Duration,
	db pqsql.Client,
	log log.Logger,
	crypto crypto.Crypto,
	gin *gin.Engine,
	stockReleaseWorker *worker.StockReleaseWorker,
) {
	stockReleaseController := controller.NewStockReleaseController(stockReleaseWorker)
	stockReleaseGroup := gin.Group("/api/stock-release")
	stockReleaseGroup.Use(middleware.JwtAuthMiddleware(env.JwtSecret))
	stockReleaseGroup.POST("/trigger", stockReleaseController.ManualRelease)
	stockReleaseGroup.GET("/status", stockReleaseController.Status)
}
