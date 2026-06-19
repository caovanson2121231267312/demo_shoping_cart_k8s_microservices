package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/caovanson/shopcaovanson/product-service/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func postgresURL(cfg *config.Config) string {
	if strings.HasPrefix(cfg.PostgresDSN, "postgres://") || strings.HasPrefix(cfg.PostgresDSN, "postgresql://") {
		return cfg.PostgresDSN
	}
	// lib/pq key=value DSN → URL for golang-migrate
	vals := map[string]string{}
	for _, part := range strings.Fields(cfg.PostgresDSN) {
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			vals[kv[0]] = kv[1]
		}
	}
	user := url.QueryEscape(vals["user"])
	pass := url.QueryEscape(vals["password"])
	host := vals["host"]
	port := vals["port"]
	dbname := vals["dbname"]
	ssl := vals["sslmode"]
	if ssl == "" {
		ssl = "disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, dbname, ssl)
}

func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	sourceURL := "file://migrations"
	dbURL := postgresURL(cfg)
	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}
	defer m.Close()

	switch *direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("migrations applied")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate down: %v", err)
		}
		log.Println("migrations rolled back")
	default:
		log.Fatalf("unknown direction: %s", *direction)
	}
}
