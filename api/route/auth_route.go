package route

import (
	"time"

	"github.com/dyaksa/warehouse/api/controller"
	"github.com/dyaksa/warehouse/bootstrap"
	"github.com/dyaksa/warehouse/infrastructure/crypto"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/dyaksa/warehouse/repository"
	"github.com/dyaksa/warehouse/usecase"
	"github.com/gin-gonic/gin"
)

func NewAuthRoute(env *bootstrap.Env, timeout time.Duration, db pqsql.Client, l log.Logger, crypto crypto.Crypto, group *gin.RouterGroup) {
	userRepository := repository.NewUserRepository(db)
	authUsecase := usecase.NewAuthUsecase(userRepository, crypto, env)

	authController := controller.AuthController{
		AuthUsecase: authUsecase,
	}

	authGroup := group.Group("/auth")
	authGroup.POST("/register", authController.Register)
	authGroup.POST("/login", authController.Login)
}
