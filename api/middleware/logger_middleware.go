package middleware

import (
	"time"

	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/gin-gonic/gin"
)

func LoggerMiddleware(l log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		duration := time.Since(start)
		l.Info("request", log.Any("method", ctx.Request.Method), log.Any("path", ctx.Request.URL.Path), log.Any("status", ctx.Writer.Status()), log.Any("duration", duration))
	}
}
