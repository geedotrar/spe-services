package configs

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	RedisAddr           string
	RedisPassword       string
	JWTSecret           string
	KafkaBrokers        []string
	KafkaTopic          string
	CheckStatusCacheTTL time.Duration
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:                getEnv("PAYMENT_PORT", "8081"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/spe?sslmode=disable"),
		RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		JWTSecret:           getEnv("JWT_SECRET", "super-secret-jwt-key"),
		KafkaBrokers:        strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaTopic:          getEnv("KAFKA_TOPIC", "payment.transaction-events"),
		CheckStatusCacheTTL: time.Duration(getEnvAsInt("CHECK_STATUS_CACHE_TTL_SECONDS", 30)) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
