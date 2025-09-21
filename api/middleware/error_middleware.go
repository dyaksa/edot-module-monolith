package middleware

import (
	"errors"

	"github.com/dyaksa/warehouse/pkg/errx"
	"github.com/dyaksa/warehouse/pkg/log"
	"github.com/gin-gonic/gin"
)

// ErrorMiddleware handles returned errors stored in context and panics, mapping them to structured JSON.
// Usage: in controllers, instead of writing JSON directly on error, call c.Error(err) and return.
func ErrorMiddleware(l log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch v := r.(type) {
				case error:
					err = v
				case string:
					err = errors.New(v)
				default:
					err = errors.New("panic")
				}
				ae := errx.ToAppError(err)
				writeError(c, ae, l)
				c.Abort()
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {
			// take the last error (most recent)
			err := c.Errors.Last().Err
			ae := errx.ToAppError(err)
			writeError(c, ae, l)
			c.Abort()
		}
	}
}

func writeError(c *gin.Context, ae *errx.AppError, l log.Logger) {
	status := errx.HTTPStatus(ae.Code)
	payload := gin.H{
		"error": gin.H{
			"code":    ae.Code,
			"message": errx.PublicMessage(ae),
		},
	}
	if len(ae.Meta) > 0 {
		payload["error"].(gin.H)["meta"] = ae.Meta
	}
	l.Error("request_error", log.Any("code", ae.Code), log.Any("message", ae.Message), log.Any("status", status), log.Any("path", c.FullPath()))
	c.JSON(status, payload)
}
