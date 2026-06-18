package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caovanson/shopcaovanson/product-service/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("cwd: %v", err)
	}

	sourceURL := fmt.Sprintf("file://%s/migrations", wd)
	m, err := migrate.New(sourceURL, cfg.PostgresDSN)
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
