package route

import (
	"time"

	"github.com/dyaksa/warehouse/bootstrap"
	"github.com/dyaksa/warehouse/infrastructure/crypto"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(env *bootstrap.Env, timeout time.Duration, db pqsql.Client, l log.Logger, crypto crypto.Crypto, r *gin.Engine) {
	publicGroup := r.Group("/api")

	NewAuthRoute(env, timeout, db, l, crypto, publicGroup)
	NewWarehouseRoute(env, timeout, db, l, crypto, publicGroup)
	NewWarehouseTransferRoute(env, timeout, db, l, crypto, publicGroup)
	NewProductRoute(env, timeout, db, l, crypto, publicGroup)
	NewShopRoute(env, timeout, db, l, crypto, publicGroup)
	NewOrderRoute(env, timeout, db, l, crypto, publicGroup)

	swaggerRoute := r.Group("/swagger")
	{
		swaggerRoute.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
