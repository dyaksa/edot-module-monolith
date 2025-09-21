package middleware

import (
	"net/http"
	"strings"

	"github.com/dyaksa/warehouse/pkg/response/response_error"
	"github.com/dyaksa/warehouse/pkg/tokenutils"
	"github.com/gin-gonic/gin"
)

func JwtAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		t := strings.Split(authHeader, " ")
		if len(t) == 2 {
			authToken := t[1]
			authorized, err := tokenutils.IsAuthorized(authToken, secret)
			if authorized {
				userID, err := tokenutils.ExtractIDFromToken(authToken, secret)
				if err != nil {
					response_error.JSON(c).Msg("Failed to extract user ID from token").Status("error").Send(http.StatusUnauthorized)
					c.Abort()
					return
				}
				c.Set("x-user-id", userID)
				c.Next()
				return
			}
			response_error.JSON(c).Msg(err.Error()).Status("error").Send(http.StatusUnauthorized)
			c.Abort()
			return
		}
		response_error.JSON(c).Msg("Not authorized").Status("error").Send(http.StatusUnauthorized)
		c.Abort()
	}
}
