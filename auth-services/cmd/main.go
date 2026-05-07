package main

import (
	"auth-services/configs"
	database "auth-services/databases"
	"auth-services/handlers"
	"auth-services/repositories"
	"auth-services/routes"
	"auth-services/services"
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := configs.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}

	merchantRepository := repositories.NewMerchantRepository(db)
	tokenRepository := repositories.NewTokenRepository(db)
	authService := services.NewAuthService(merchantRepository, tokenRepository, cfg.JWTSecret, int64(cfg.JWTTTL.Minutes()))
	authHandler := handlers.NewAuthHandler(authService)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	engine.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok", "service": "auth-service"})
	})

	routes.Register(engine, authHandler, redisClient, cfg.RateLimitMaxAttempt, cfg.RateLimitWindow)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("auth service listening", "port", cfg.Port)

	if err := engine.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run auth service: %v", err)
	}

}
