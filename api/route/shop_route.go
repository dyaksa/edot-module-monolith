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

func NewShopRoute(env *bootstrap.Env, timeout time.Duration, db pqsql.Client, l log.Logger, crypto crypto.Crypto, group *gin.RouterGroup) {
	jwtMiddleware := middleware.JwtAuthMiddleware(env.JwtSecret)
	shopRepository := repository.NewShopRepository(db)
	shopUsecase := usecase.NewShopUsecase(shopRepository)

	shopController := controller.ShopController{
		ShopUsecase: shopUsecase,
	}

	shopGroup := group.Group("/shop", jwtMiddleware)
	shopGroup.POST("/create", shopController.Create)
	shopGroup.GET("/retrieve", shopController.Retrieve)
	shopGroup.PUT("/update", shopController.Update)
	shopGroup.DELETE("/delete", shopController.Delete)
}
