package bootstrap

import (
	"context"

	"github.com/dyaksa/warehouse/infrastructure/crypto"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/dyaksa/warehouse/pkg/log/logrus"
	"github.com/dyaksa/warehouse/pkg/validationutils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Application struct {
	Env      *Env
	App      *gin.Engine
	Log      log.Logger
	Postgres pqsql.Client
	Crypto   crypto.Crypto
}

func App(ctx context.Context) *Application {
	app := &Application{
		Env: NewEnv(ctx),
		App: gin.Default(),
	}

	ll, err := logrus.New(
		logrus.WithLevel("info"),
		logrus.WithJSONFormatter(),
	)

	if err != nil {
		panic(err)
	}

	app.Log = ll
	app.Postgres = NewPostgres(app.Env, app.Log)
	app.Crypto = NewDerivaleCrypto(app.Log)

	return app
}

func (app *Application) CloseConnection() {
	CloseConnection(app.Postgres, app.Log)
}

func (app *Application) CustomValidation() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = validationutils.IdentifierValidator(v)
	}
}
