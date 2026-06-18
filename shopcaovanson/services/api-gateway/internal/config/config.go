package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	RedisURL           string
	JWTPublicKeyPEM    string
	AuthServiceURL     string
	ProductServiceURL  string
	OrderServiceURL    string
	ChatServiceURL         string
	NotificationServiceURL string
	CORSAllowedOrigins string
	RateLimitPerMinute int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	publicKey := os.Getenv("JWT_PUBLIC_KEY")
	if publicKey == "" {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY is required")
	}

	rateLimit, err := strconv.Atoi(getEnv("RATE_LIMIT_PER_MINUTE", "100"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_PER_MINUTE: %w", err)
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTPublicKeyPEM:    publicKey,
		AuthServiceURL:     getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		ProductServiceURL:  getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
		OrderServiceURL:    getEnv("ORDER_SERVICE_URL", "http://localhost:8083"),
		ChatServiceURL:         getEnv("CHAT_SERVICE_URL", "http://localhost:8084"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8085"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "https://shopcaovanson.xyz"),
		RateLimitPerMinute: rateLimit,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
