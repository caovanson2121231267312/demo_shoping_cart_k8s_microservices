package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	MongoURI     string
	MongoDB      string
	RedisURL     string
	JWTPublicKey string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDB := os.Getenv("MONGODB_DATABASE")
	if mongoDB == "" {
		mongoDB = "shop_chat"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	return &Config{
		Port:         port,
		MongoURI:     mongoURI,
		MongoDB:      mongoDB,
		RedisURL:     redisURL,
		JWTPublicKey: os.Getenv("JWT_PUBLIC_KEY"),
	}
}

func GetEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return n
}
