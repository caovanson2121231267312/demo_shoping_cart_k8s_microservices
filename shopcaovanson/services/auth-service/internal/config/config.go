package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	FrontendURL      string
	Port             string
	DatabaseURL      string
	RedisURL         string
	KafkaBrokers     string
	JWTPrivateKeyPEM string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	BcryptCost       int
	MinIOEnabled     bool
	MinIOEndpoint    string
	MinIOAccessKey   string
	MinIOSecretKey   string
	MinIOBucket      string
	MinIOUseSSL      bool
	MinIORegion      string
	MaxAvatarBytes   int64
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	accessMinutes, err := strconv.Atoi(getEnv("ACCESS_TOKEN_TTL_MINUTES", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_TTL_MINUTES: %w", err)
	}

	refreshDays, err := strconv.Atoi(getEnv("REFRESH_TOKEN_TTL_DAYS", "7"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_TTL_DAYS: %w", err)
	}

	bcryptCost, err := strconv.Atoi(getEnv("BCRYPT_COST", "12"))
	if err != nil {
		return nil, fmt.Errorf("invalid BCRYPT_COST: %w", err)
	}

	privateKey := os.Getenv("JWT_PRIVATE_KEY")
	if privateKey == "" {
		return nil, fmt.Errorf("JWT_PRIVATE_KEY is required")
	}
	privateKey = strings.ReplaceAll(privateKey, "\\n", "\n")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "shopcaovanson")
		pass := os.Getenv("DB_PASSWORD")
		name := getEnv("DB_NAME", "auth_db")
		ssl := getEnv("DB_SSLMODE", "disable")
		if pass != "" {
			dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, name, ssl)
		} else {
			dbURL = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", host, port, user, name, ssl)
		}
	}
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	maxAvatarMB, err := strconv.Atoi(getEnv("MAX_AVATAR_MB", "2"))
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_AVATAR_MB: %w", err)
	}

	return &Config{
		Port:             getEnv("PORT", "8081"),
		DatabaseURL:      dbURL,
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379/0"),
		KafkaBrokers:     getEnv("KAFKA_BROKERS", "localhost:9092"),
		JWTPrivateKeyPEM: privateKey,
		AccessTokenTTL:   time.Duration(accessMinutes) * time.Minute,
		RefreshTokenTTL:  time.Duration(refreshDays) * 24 * time.Hour,
		BcryptCost:       bcryptCost,
		FrontendURL:      getEnv("FRONTEND_URL", "http://localhost:3000"),
		MinIOEnabled:     strings.EqualFold(getEnv("MINIO_ENABLED", "false"), "true"),
		MinIOEndpoint:    getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:   os.Getenv("MINIO_ACCESS_KEY"),
		MinIOSecretKey:   os.Getenv("MINIO_SECRET_KEY"),
		MinIOBucket:      getEnv("MINIO_BUCKET", "shopcaovanson"),
		MinIOUseSSL:      strings.EqualFold(getEnv("MINIO_USE_SSL", "false"), "true"),
		MinIORegion:      getEnv("MINIO_REGION", "us-east-1"),
		MaxAvatarBytes:   int64(maxAvatarMB) * 1024 * 1024,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
