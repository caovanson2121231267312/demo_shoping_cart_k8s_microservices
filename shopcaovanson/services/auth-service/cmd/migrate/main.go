package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/shopcaovanson/auth-service/internal/config"
)

func main() {
	direction := flag.String("direction", "", "migration direction: up, down, or force")
	steps := flag.Int("steps", 0, "number of migration steps (0 = all)")
	forceVersion := flag.Int("version", -1, "version for force (clears dirty flag)")
	flag.Parse()

	dir := *direction
	args := flag.Args()
	if dir == "" && len(args) > 0 {
		dir = args[0]
		args = args[1:]
	}
	if dir == "" {
		dir = "up"
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	dbURL := cfg.DatabaseURL

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
		case "force":
			ver := *forceVersion
			if ver < 0 && len(args) > 0 {
				ver, runErr = strconv.Atoi(args[0])
			}
			if runErr != nil {
				done <- fmt.Errorf("force version: %w", runErr)
				return
			}
			if ver < 0 {
				done <- fmt.Errorf("force requires a version, e.g. migrate force 7")
				return
			}
			runErr = m.Force(ver)
		case "version":
			v, dirty, vErr := m.Version()
			if vErr != nil && vErr != migrate.ErrNilVersion {
				done <- vErr
				return
			}
			log.Printf("current version=%d dirty=%v", v, dirty)
			done <- nil
			return
		default:
			runErr = fmt.Errorf("invalid direction: %s (use up|down|force|version)", dir)
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
