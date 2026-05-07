package routes

import (
	"time"

	"auth-services/handlers"
	"auth-services/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func Register(engine *gin.Engine, authHandler *handlers.AuthHandler, redisClient *redis.Client, rateLimitMax int, rateLimitWindow time.Duration) {
	api := engine.Group("/api/v1")
	api.POST("/auth/token", middlewares.RateLimitMiddleware(redisClient, rateLimitMax, rateLimitWindow), authHandler.IssueToken)
}
