package route

import (
	"time"

	"github.com/dyaksa/warehouse/api/controller"
	"github.com/dyaksa/warehouse/api/middleware"
	"github.com/dyaksa/warehouse/bootstrap"
	"github.com/dyaksa/warehouse/infrastructure/crypto"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/dyaksa/warehouse/repository"
	"github.com/dyaksa/warehouse/usecase"
	"github.com/gin-gonic/gin"
)

func NewWarehouseTransferRoute(env *bootstrap.Env, timeout time.Duration, db pqsql.Client, l log.Logger, crypto crypto.Crypto, group *gin.RouterGroup) {
	jwtMiddleware := middleware.JwtAuthMiddleware(env.JwtSecret)

	// Initialize repositories
	warehouseTransferRepo := repository.NewWarehouseTransferRepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)
	productStockRepo := repository.NewProductStockRepository(db)
	movementRepo := repository.NewMovementRepository(db)

	// Initialize usecase
	warehouseTransferUsecase := usecase.NewWarehouseTransferUsecase(
		db.Database(),
		warehouseTransferRepo,
		warehouseRepo,
		productStockRepo,
		movementRepo,
	)

	// Initialize controller
	warehouseTransferController := controller.WarehouseTransferController{
		TransferUsecase: warehouseTransferUsecase,
	}

	// Create route group
	transferGroup := group.Group("/transfers", jwtMiddleware)
	transferGroup.POST("/", warehouseTransferController.CreateTransfer)
	transferGroup.GET("/:id", warehouseTransferController.GetTransfer)
	transferGroup.PUT("/:id/status", warehouseTransferController.UpdateTransferStatus)
	transferGroup.POST("/:id/execute", warehouseTransferController.ExecuteTransfer)
	transferGroup.GET("/warehouse/:warehouse_id", warehouseTransferController.GetTransfersByWarehouse)
}
