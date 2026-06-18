package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	Env               string
	PostgresDSN       string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	CartTTLHours      int
	ProductServiceURL string
	KafkaBrokers      []string
}

func Load() (*Config, error) {
	port := getEnv("APP_PORT", getEnv("PORT", "8083"))
	pgHost := getEnv("POSTGRES_HOST", getEnv("DB_HOST", "localhost"))
	pgPort := getEnv("POSTGRES_PORT", getEnv("DB_PORT", "5432"))
	pgUser := getEnv("POSTGRES_USER", getEnv("DB_USER", "shop"))
	pgPass := getEnv("POSTGRES_PASSWORD", os.Getenv("DB_PASSWORD"))
	if pgPass == "" {
		pgPass = "shop_secret"
	}
	pgDB := getEnv("POSTGRES_DB", getEnv("DB_NAME", "order_db"))
	pgSSL := getEnv("POSTGRES_SSLMODE", getEnv("DB_SSLMODE", "disable"))

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		pgHost, pgPort, pgUser, pgPass, pgDB, pgSSL,
	)

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPass := getEnv("REDIS_PASSWORD", os.Getenv("REDIS_PASSWORD"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	cartTTL, _ := strconv.Atoi(getEnv("CART_TTL_HOURS", "720"))

	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	return &Config{
		Port:              port,
		Env:               getEnv("APP_ENV", "development"),
		PostgresDSN:       dsn,
		RedisAddr:         fmt.Sprintf("%s:%s", redisHost, redisPort),
		RedisPassword:     redisPass,
		RedisDB:           redisDB,
		CartTTLHours:      cartTTL,
		ProductServiceURL: strings.TrimRight(getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"), "/"),
		KafkaBrokers:      brokers,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
