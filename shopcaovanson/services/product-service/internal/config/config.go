package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	Env               string
	PostgresDSN       string
	MongoURI          string
	MongoDB           string
	ElasticsearchURL  string
	ElasticsearchIndex string
	KafkaBrokers      []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port := getEnv("APP_PORT", getEnv("PORT", "8082"))
	pgHost := getEnv("POSTGRES_HOST", getEnv("DB_HOST", "localhost"))
	pgPort := getEnv("POSTGRES_PORT", getEnv("DB_PORT", "5432"))
	pgUser := getEnv("POSTGRES_USER", getEnv("DB_USER", "shop"))
	pgPass := getEnv("POSTGRES_PASSWORD", os.Getenv("DB_PASSWORD"))
	if pgPass == "" {
		pgPass = "shop_secret"
	}
	pgDB := getEnv("POSTGRES_DB", getEnv("DB_NAME", "product_db"))
	pgSSL := getEnv("POSTGRES_SSLMODE", getEnv("DB_SSLMODE", "disable"))

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		pgHost, pgPort, pgUser, pgPass, pgDB, pgSSL,
	)

	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	return &Config{
		Port:               port,
		Env:                getEnv("APP_ENV", "development"),
		PostgresDSN:        dsn,
		MongoURI:           getEnv("MONGO_URI", fmt.Sprintf("mongodb://%s:%s", getEnv("MONGO_HOST", "localhost"), getEnv("MONGO_PORT", "27017"))),
		MongoDB:            getEnv("MONGO_DB", getEnv("DB_NAME", "shop")),
		ElasticsearchURL:   getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		ElasticsearchIndex: getEnv("ELASTICSEARCH_INDEX", "products"),
		KafkaBrokers:       brokers,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
