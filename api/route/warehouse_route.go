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

func NewWarehouseRoute(env *bootstrap.Env, timeout time.Duration, db pqsql.Client, l log.Logger, crypto crypto.Crypto, group *gin.RouterGroup) {
	jwtMiddleware := middleware.JwtAuthMiddleware(env.JwtSecret)
	wareHouseRepository := repository.NewWarehouseRepository(db)
	warehouseTransferRepository := repository.NewWarehouseTransferRepository(db)
	wareHouseUsecase := usecase.NewWarehouseUsecase(wareHouseRepository, warehouseTransferRepository)

	wareHouseController := controller.WarehouseController{
		WarehouseUsecase: wareHouseUsecase,
	}

	warehouseGroup := group.Group("/warehouse", jwtMiddleware)
	warehouseGroup.POST("/create", wareHouseController.Create)
	warehouseGroup.GET("/:id", wareHouseController.Retrieve)
	warehouseGroup.PUT("/:id", wareHouseController.Update)
	warehouseGroup.DELETE("/:id", wareHouseController.Delete)
	warehouseGroup.PUT("/:id/status", wareHouseController.SetActive)
	warehouseGroup.GET("/shop/:shop_id", wareHouseController.GetByShop)
}
