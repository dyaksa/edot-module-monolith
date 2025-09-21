package route

import (
	"time"

	"github.com/dyaksa/warehouse/bootstrap"
	"github.com/dyaksa/warehouse/infrastructure/crypto"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/gin-gonic/gin"
)

func NewUserRoute(env *bootstrap.Env, timeout time.Duration, db pqsql.Client, l log.Logger, crypto crypto.Crypto, group *gin.RouterGroup) {
}
