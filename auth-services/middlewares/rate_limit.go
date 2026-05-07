package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"auth-services/helpers"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(client *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := fmt.Sprintf("auth:ratelimit:%s", ctx.ClientIP())

		count, err := client.Incr(context.Background(), key).Result()
		if err != nil {
			helpers.JSONError(ctx, http.StatusInternalServerError, "99", "rate limiter unavailable")
			ctx.Abort()
			return
		}

		if count == 1 {
			client.Expire(context.Background(), key, window)
		}

		if int(count) > limit {
			client.Expire(context.Background(), key, window)

			helpers.JSONError(ctx, http.StatusTooManyRequests, "96", "too many authentication attempts")

			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
