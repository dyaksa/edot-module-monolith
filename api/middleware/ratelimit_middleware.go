package middleware

import (
	"fmt"
	"time"

	"github.com/dyaksa/warehouse/pkg/ratelimit"
	"github.com/gin-gonic/gin"
)

type Options func(*ratelimit.Options)

func keyFunc(ctx *gin.Context) string {
	return ctx.ClientIP()
}

func errorHandler(ctx *gin.Context, info ratelimit.Info) {
	ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
	ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", info.RemainingHits))
	ctx.Header("X-RateLimit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))
	ctx.String(429, "Too Many Requests")
}

func beforeResponse(ctx *gin.Context, info ratelimit.Info) {
	ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", info.Limit))
	ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", info.RemainingHits))
	ctx.Header("X-RateLimit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))
}

func RateLimit(store ratelimit.Store) gin.HandlerFunc {
	options := &ratelimit.Options{
		KeyFunc:        keyFunc,
		ErrorHandler:   errorHandler,
		BeforeResponse: beforeResponse,
	}

	return func(ctx *gin.Context) {
		key := options.KeyFunc(ctx)
		info := store.Limit(key, ctx)
		options.BeforeResponse(ctx, info)

		if ctx.IsAborted() {
			return
		}

		if info.RateLimited {
			options.ErrorHandler(ctx, info)
			ctx.Abort()
		} else {
			ctx.Next()
		}
	}
}

func RateLimitDefault() gin.HandlerFunc {
	return RateLimit(ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: 100,
	}))
}

// RateLimitRedisCustom creates a Redis-based rate limiter with custom settings
func RateLimitMiddleware(rate time.Duration, limit uint, prefix string) gin.HandlerFunc {
	return RateLimit(ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  rate,
		Limit: limit,
	}))
}
