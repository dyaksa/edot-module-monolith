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

func NewProductRoute(env *bootstrap.Env, timeout time.Duration, db pqsql.Client, l log.Logger, crypto crypto.Crypto, group *gin.RouterGroup) {
	jwtMiddleware := middleware.JwtAuthMiddleware(env.JwtSecret)
	productRepository := repository.NewProductRepository(db)
	productStockRepository := repository.NewProductStockRepository(db)
	productUsecase := usecase.NewProductUsecase(productRepository, productStockRepository)

	productController := controller.ProductController{
		ProductUsecase: productUsecase,
	}

	groupProduct := group.Group("/product")

	groupProduct.POST("/create", jwtMiddleware, productController.Create)
	groupProduct.GET("/list", jwtMiddleware, productController.RetrieveAll)
}
