package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"payment-services/configs"
	"payment-services/databases"
	"payment-services/handlers"
	"payment-services/repositories"
	"payment-services/routes"
	"payment-services/services"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := configs.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// connect db
	db, err := databases.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// connect redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}

	// setup kafka producer and consumer
	kafkaConfig := sarama.NewConfig()
	// ensure producer finish sending messages
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Return.Errors = true
	kafkaConfig.Version = sarama.V3_7_0_0

	var eventService *services.EventService
	producer, err := sarama.NewAsyncProducer(cfg.KafkaBrokers, kafkaConfig)
	if err != nil {
		logger.Warn("kafka producer not connected, running without async events", "error", err)
	} else {
		defer producer.AsyncClose()
		transactionEventLogRepository := repositories.NewTransactionEventLogRepository(db)
		eventService = services.NewEventService(producer, cfg.KafkaTopic, logger, transactionEventLogRepository)
		if err := eventService.StartTransactionConsumer(context.Background(), cfg.KafkaBrokers); err != nil {
			logger.Warn("transaction consumer not started", "error", err)
		}
	}

	merchantRepository := repositories.NewMerchantRepository(db)
	transactionRepository := repositories.NewTransactionRepository(db)
	paymentService := services.NewPaymentService(merchantRepository, transactionRepository, redisClient, eventService, cfg.CheckStatusCacheTTL)
	paymentHandler := handlers.NewPaymentHandler(paymentService)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	engine.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok", "service": "payment-service"})
	})
	routes.Register(engine, paymentHandler, paymentService, cfg.JWTSecret)

	logger.Info("payment service listening", "port", cfg.Port)
	if err := engine.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run payment service: %v", err)
	}
}
