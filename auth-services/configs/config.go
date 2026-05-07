package configs

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	RedisAddr           string
	RedisPassword       string
	JWTSecret           string
	JWTTTL              time.Duration
	RateLimitMaxAttempt int
	RateLimitWindow     time.Duration
}

func Load() Config {
	_ = godotenv.Load()

	jwtTTLMinutes := getEnvAsInt("JWT_TTL_MINUTES", 15)
	rateLimitMax := getEnvAsInt("RATE_LIMIT_MAX_ATTEMPTS", 5)
	rateLimitWindow := getEnvAsInt("RATE_LIMIT_WINDOW_MINUTES", 10)

	cfg := Config{
		Port:                getEnv("AUTH_PORT", "9090"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/spe?sslmode=disable"),
		RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		JWTSecret:           getEnv("JWT_SECRET", "super-secret-jwt-key"),
		JWTTTL:              time.Duration(jwtTTLMinutes) * time.Minute,
		RateLimitMaxAttempt: rateLimitMax,
		RateLimitWindow:     time.Duration(rateLimitWindow) * time.Minute,
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	return cfg
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
