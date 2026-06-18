package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	direction := flag.String("direction", "", "migration direction: up or down")
	steps := flag.Int("steps", 0, "number of migration steps (0 = all)")
	flag.Parse()

	dir := *direction
	if dir == "" && flag.NArg() > 0 {
		dir = flag.Arg(0)
	}
	if dir == "" {
		dir = "up"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://migrations"
	}

	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		log.Fatalf("create migrator: %v", err)
	}
	defer m.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		var runErr error
		switch dir {
		case "up":
			if *steps > 0 {
				runErr = m.Steps(*steps)
			} else {
				runErr = m.Up()
			}
		case "down":
			if *steps > 0 {
				runErr = m.Steps(-*steps)
			} else {
				runErr = m.Down()
			}
		default:
			runErr = fmt.Errorf("invalid direction: %s", dir)
		}
		done <- runErr
	}()

	select {
	case <-ctx.Done():
		log.Fatal("migration timed out")
	case err := <-done:
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration failed: %v", err)
		}
	}

	version, dirty, _ := m.Version()
	log.Printf("migration %s complete (version=%d dirty=%v)", dir, version, dirty)
}
