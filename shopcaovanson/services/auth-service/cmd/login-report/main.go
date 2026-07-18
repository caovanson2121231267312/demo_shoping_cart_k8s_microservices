package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/shopcaovanson/auth-service/internal/config"
	"github.com/shopcaovanson/auth-service/internal/repository"
	"github.com/shopcaovanson/auth-service/internal/service"
	"github.com/shopcaovanson/auth-service/internal/storage"
)

func main() {
	dateFlag := flag.String("date", "", "Report date YYYY-MM-DD (VN timezone). Default: yesterday")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if !cfg.MinIOEnabled {
		log.Fatal("MINIO_ENABLED must be true to generate login reports")
	}

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	store, err := storage.NewAvatarStorage(cfg)
	if err != nil {
		log.Fatalf("minio: %v", err)
	}

	svc := service.NewLoginHistoryService(
		repository.NewLoginHistoryRepository(db),
		repository.NewLoginReportRepository(db),
		store,
	)

	day := service.YesterdayVN(time.Now())
	if *dateFlag != "" {
		parsed, err := time.ParseInLocation("2006-01-02", *dateFlag, service.VietnamLocation())
		if err != nil {
			log.Fatalf("invalid -date: %v", err)
		}
		day = parsed
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	report, err := svc.GenerateDailyReport(ctx, day)
	if err != nil {
		log.Fatalf("generate report: %v", err)
	}

	fmt.Printf("login report ready: date=%s file=%s key=%s total=%d success=%d failure=%d\n",
		day.Format("2006-01-02"), report.FileName, report.ObjectKey,
		report.TotalLogins, report.SuccessCount, report.FailureCount)
	os.Exit(0)
}
